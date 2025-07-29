package services

import (
   "testing"
   "time"

   "github.com/google/uuid"
)

// TestNotificationService_PublishSubscribe verifies notifications are received by subscribers.
func TestNotificationService_PublishSubscribe(t *testing.T) {
   svc := NewNotificationService()
   userID := uuid.New()
   ch, cancel := svc.Subscribe(userID)
   defer cancel()

   notif := Notification{
       ID:        uuid.New(),
       Type:      "test",
       Payload:   "payload",
       Timestamp: time.Now(),
   }
   svc.Publish(userID, notif)

   select {
   case got := <-ch:
       if got.ID != notif.ID || got.Type != notif.Type {
           t.Errorf("got %+v, want %+v", got, notif)
       }
   case <-time.After(time.Second):
       t.Fatal("timeout waiting for notification")
   }
}
