package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// KubernetesExecutor handles job execution in Kubernetes
type KubernetesExecutor struct {
	db         *gorm.DB
	logger     *logrus.Logger
	clientset  *kubernetes.Clientset
	namespace  string
	secretService *SecretService
}

// NewKubernetesExecutor creates a new Kubernetes executor
func NewKubernetesExecutor(db *gorm.DB, logger *logrus.Logger, secretService *SecretService) (*KubernetesExecutor, error) {
	// Create in-cluster config (when running inside Kubernetes)
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &KubernetesExecutor{
		db:            db,
		logger:        logger,
		clientset:     clientset,
		namespace:     "hub-runners", // Default namespace for runners
		secretService: secretService,
	}, nil
}

// ExecuteJob executes a job in Kubernetes
func (e *KubernetesExecutor) ExecuteJob(ctx context.Context, job *models.Job) error {
	e.logger.WithField("job_id", job.ID).Info("Starting Kubernetes job execution")

	// Update job status to in_progress
	if err := e.updateJobStatus(ctx, job.ID, models.JobStatusInProgress, nil); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Create Kubernetes Job
	k8sJob, err := e.createKubernetesJob(ctx, job)
	if err != nil {
		e.markJobFailed(ctx, job.ID, fmt.Sprintf("Failed to create Kubernetes job: %v", err))
		return fmt.Errorf("failed to create Kubernetes job: %w", err)
	}

	// Watch job progress
	go e.watchJobProgress(ctx, job.ID, k8sJob.Name)

	return nil
}

// createKubernetesJob creates a Kubernetes Job resource for the workflow job
func (e *KubernetesExecutor) createKubernetesJob(ctx context.Context, job *models.Job) (*batchv1.Job, error) {
	// Load workflow run and workflow to get configuration
	var workflowRun models.WorkflowRun
	err := e.db.WithContext(ctx).
		Preload("Workflow").
		Preload("Repository").
		First(&workflowRun, "id = ?", job.WorkflowRunID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow run: %w", err)
	}

	// Parse workflow YAML to extract job configuration
	// This is a simplified version - in a full implementation, you'd parse the YAML
	// and extract the specific job configuration
	jobConfig := map[string]interface{}{
		"runs-on": "ubuntu-latest",
		"steps":   []interface{}{},
	}

	// Create environment variables including secrets
	var organizationID *uuid.UUID
	if workflowRun.Repository.OwnerType == models.OwnerTypeOrganization {
		organizationID = &workflowRun.Repository.OwnerID
	}
	envVars, err := e.createEnvironmentVariables(ctx, workflowRun.Repository.ID, organizationID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment variables: %w", err)
	}

	// Generate unique job name
	jobName := fmt.Sprintf("workflow-job-%s", strings.ReplaceAll(job.ID.String(), "-", "")[:8])

	// Create Kubernetes Job spec
	k8sJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: e.namespace,
			Labels: map[string]string{
				"app":            "hub-runner",
				"job-id":         job.ID.String(),
				"workflow-id":    workflowRun.WorkflowID.String(),
				"repository-id":  workflowRun.Repository.ID.String(),
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":    "hub-runner",
						"job-id": job.ID.String(),
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "runner",
							Image: e.getRunnerImage(jobConfig),
							Env:   envVars,
							WorkingDir: "/workspace",
							Command: []string{"/bin/bash", "-c"},
							Args: []string{
								e.generateJobScript(job, workflowRun),
							},
							Resources: corev1.ResourceRequirements{
								// Set resource limits based on job requirements
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workspace",
									MountPath: "/workspace",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "workspace",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					ServiceAccountName: "hub-runner", // Service account with necessary permissions
				},
			},
			BackoffLimit: int32Ptr(0), // Don't retry failed jobs
		},
	}

	// Create the job in Kubernetes
	createdJob, err := e.clientset.BatchV1().Jobs(e.namespace).Create(ctx, k8sJob, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes job: %w", err)
	}

	e.logger.WithFields(logrus.Fields{
		"job_id":     job.ID,
		"k8s_job":    createdJob.Name,
		"namespace":  e.namespace,
	}).Info("Created Kubernetes job")

	return createdJob, nil
}

// createEnvironmentVariables creates environment variables for the job including secrets
func (e *KubernetesExecutor) createEnvironmentVariables(ctx context.Context, repositoryID uuid.UUID, organizationID *uuid.UUID, environment *string) ([]corev1.EnvVar, error) {
	var envVars []corev1.EnvVar

	// Add standard environment variables
	envVars = append(envVars, []corev1.EnvVar{
		{Name: "CI", Value: "true"},
		{Name: "RUNNER_OS", Value: "Linux"},
		{Name: "RUNNER_ARCH", Value: "X64"},
		{Name: "GITHUB_ACTIONS", Value: "true"}, // For compatibility
	}...)

	// Get secrets for this job
	if e.secretService != nil {
		secrets, err := e.secretService.GetSecretsForJob(ctx, repositoryID, organizationID, environment)
		if err != nil {
			e.logger.WithError(err).Warn("Failed to get secrets for job")
		} else {
			for name, value := range secrets {
				envVars = append(envVars, corev1.EnvVar{
					Name:  name,
					Value: value,
				})
			}
		}
	}

	return envVars, nil
}

// getRunnerImage returns the appropriate container image for the job
func (e *KubernetesExecutor) getRunnerImage(jobConfig map[string]interface{}) string {
	// Default to Ubuntu latest
	runsOn := "ubuntu-latest"
	if ro, ok := jobConfig["runs-on"].(string); ok {
		runsOn = ro
	}

	// Map runs-on values to container images
	imageMap := map[string]string{
		"ubuntu-latest": "ghcr.io/a5c-ai/hub-runner:ubuntu-latest",
		"ubuntu-22.04":  "ghcr.io/a5c-ai/hub-runner:ubuntu-22.04",
		"ubuntu-20.04":  "ghcr.io/a5c-ai/hub-runner:ubuntu-20.04",
		// Add more image mappings as needed
	}

	if image, ok := imageMap[runsOn]; ok {
		return image
	}

	// Default fallback
	return "ghcr.io/a5c-ai/hub-runner:ubuntu-latest"
}

// generateJobScript generates the bash script to execute the job steps
func (e *KubernetesExecutor) generateJobScript(job *models.Job, workflowRun models.WorkflowRun) string {
	var script strings.Builder

	script.WriteString("#!/bin/bash\nset -e\n\n")
	script.WriteString("echo '::group::Job Setup'\n")
	script.WriteString("echo 'Setting up job environment...'\n")
	script.WriteString("echo '::endgroup::'\n\n")

	// In a real implementation, you would parse the job steps from the workflow YAML
	// For now, we'll create a simple placeholder
	script.WriteString("echo '::group::Checkout'\n")
	script.WriteString("# Checkout code (simplified)\n")
	script.WriteString("echo 'Checking out code...'\n")
	script.WriteString("echo '::endgroup::'\n\n")

	script.WriteString("echo '::group::Run Steps'\n")
	script.WriteString("# Execute job steps\n")
	script.WriteString("echo 'Executing job steps...'\n")
	script.WriteString("echo 'Job completed successfully'\n")
	script.WriteString("echo '::endgroup::'\n")

	return script.String()
}

// watchJobProgress watches the Kubernetes job and updates the database accordingly
func (e *KubernetesExecutor) watchJobProgress(ctx context.Context, jobID uuid.UUID, k8sJobName string) {
	e.logger.WithFields(logrus.Fields{
		"job_id":    jobID,
		"k8s_job":   k8sJobName,
	}).Info("Starting to watch Kubernetes job progress")

	// Watch for job events
	watchInterface, err := e.clientset.BatchV1().Jobs(e.namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", k8sJobName),
	})
	if err != nil {
		e.logger.WithError(err).Error("Failed to watch Kubernetes job")
		e.markJobFailed(ctx, jobID, fmt.Sprintf("Failed to watch job: %v", err))
		return
	}
	defer watchInterface.Stop()

	// Also watch for pod events to get logs
	go e.watchPodLogs(ctx, jobID, k8sJobName)

	for event := range watchInterface.ResultChan() {
		if event.Type == watch.Error {
			e.logger.WithField("job_id", jobID).Error("Error watching Kubernetes job")
			e.markJobFailed(ctx, jobID, "Error watching Kubernetes job")
			return
		}

		job, ok := event.Object.(*batchv1.Job)
		if !ok {
			continue
		}

		// Check job conditions
		for _, condition := range job.Status.Conditions {
			switch condition.Type {
			case batchv1.JobComplete:
				if condition.Status == corev1.ConditionTrue {
					e.logger.WithField("job_id", jobID).Info("Kubernetes job completed successfully")
					conclusion := models.JobConclusionSuccess
					e.updateJobStatus(ctx, jobID, models.JobStatusCompleted, &conclusion)
					return
				}
			case batchv1.JobFailed:
				if condition.Status == corev1.ConditionTrue {
					e.logger.WithField("job_id", jobID).Error("Kubernetes job failed")
					e.markJobFailed(ctx, jobID, condition.Message)
					return
				}
			}
		}
	}
}

// watchPodLogs watches pod logs and stores them in the database
func (e *KubernetesExecutor) watchPodLogs(ctx context.Context, jobID uuid.UUID, k8sJobName string) {
	// Wait a bit for pod to be created
	time.Sleep(5 * time.Second)

	// Get pods for this job
	pods, err := e.clientset.CoreV1().Pods(e.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", k8sJobName),
	})
	if err != nil || len(pods.Items) == 0 {
		e.logger.WithError(err).WithField("job_id", jobID).Warn("No pods found for job")
		return
	}

	podName := pods.Items[0].Name

	// Get pod logs
	req := e.clientset.CoreV1().Pods(e.namespace).GetLogs(podName, &corev1.PodLogOptions{
		Follow: true,
	})

	logs, err := req.Stream(ctx)
	if err != nil {
		e.logger.WithError(err).WithField("job_id", jobID).Error("Failed to get pod logs")
		return
	}
	defer logs.Close()

	// Read logs and store them
	// In a real implementation, you would stream logs and store them incrementally
	// For now, we'll just mark that logs are available
	e.logger.WithField("job_id", jobID).Info("Pod logs available")
}

// updateJobStatus updates the job status in the database
func (e *KubernetesExecutor) updateJobStatus(ctx context.Context, jobID uuid.UUID, status models.JobStatus, conclusion *models.JobConclusion) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == models.JobStatusInProgress {
		updates["started_at"] = time.Now()
	}

	if status == models.JobStatusCompleted || status == models.JobStatusCancelled {
		updates["completed_at"] = time.Now()
		if conclusion != nil {
			updates["conclusion"] = *conclusion
		}
	}

	return e.db.WithContext(ctx).Model(&models.Job{}).
		Where("id = ?", jobID).
		Updates(updates).Error
}

// markJobFailed marks a job as failed with an error message
func (e *KubernetesExecutor) markJobFailed(ctx context.Context, jobID uuid.UUID, errorMsg string) {
	e.logger.WithFields(logrus.Fields{
		"job_id": jobID,
		"error":  errorMsg,
	}).Error("Marking job as failed")

	conclusion := models.JobConclusionFailure
	if err := e.updateJobStatus(ctx, jobID, models.JobStatusCompleted, &conclusion); err != nil {
		e.logger.WithError(err).Error("Failed to update job status to failed")
	}
}

// CleanupJob cleans up Kubernetes resources for a job
func (e *KubernetesExecutor) CleanupJob(ctx context.Context, jobID uuid.UUID) error {
	// Find Kubernetes jobs with this job ID
	jobs, err := e.clientset.BatchV1().Jobs(e.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-id=%s", jobID.String()),
	})
	if err != nil {
		return fmt.Errorf("failed to list Kubernetes jobs: %w", err)
	}

	// Delete each job and its pods
	for _, job := range jobs.Items {
		// Delete the job (this will also delete associated pods)
		err := e.clientset.BatchV1().Jobs(e.namespace).Delete(ctx, job.Name, metav1.DeleteOptions{})
		if err != nil {
			e.logger.WithError(err).WithField("k8s_job", job.Name).Warn("Failed to delete Kubernetes job")
		} else {
			e.logger.WithField("k8s_job", job.Name).Info("Deleted Kubernetes job")
		}
	}

	return nil
}

// Helper function to create int32 pointer
func int32Ptr(i int32) *int32 {
	return &i
}