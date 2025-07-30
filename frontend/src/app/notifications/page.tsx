'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Avatar } from '@/components/ui/Avatar';
import { Button } from '@/components/ui/Button';
import api from '@/lib/api';
import Link from 'next/link';
import { createErrorHandler } from '@/lib/utils';
import { usePWAContext } from '@/components/providers/PWAProvider';

interface Notification {
  id: string;
  type: 'issue' | 'pull_request' | 'mention' | 'security_alert' | 'repository_invite';
  title: string;
  body?: string;
  repository?: {
    id: string;
    name: string;
    full_name: string;
    owner: {
      username: string;
      avatar_url?: string;
    };
  };
  subject: {
    title: string;
    url?: string;
    type: string;
  };
  reason: 'subscribed' | 'mentioned' | 'assigned' | 'author' | 'comment' | 'invitation' | 'security_alert';
  unread: boolean;
  updated_at: string;
  last_read_at?: string;
}

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<'all' | 'unread' | 'participating'>('unread');
  const [selectedNotifications, setSelectedNotifications] = useState<string[]>([]);
  const { subscribeToPushNotifications, unsubscribeFromPushNotifications, isOnline } = usePWAContext();
  const [pushSubscribed, setPushSubscribed] = useState(false);

  const fetchNotifications = async (currentFilter = filter) => {
    const handleError = createErrorHandler(setError, setLoading);
    
    const operation = async () => {
      const response = await api.get(`/notifications?filter=${currentFilter}`);
      return response.data;
    };

    const result = await handleError(operation);
    if (result) {
      setNotifications(result);
    }
  };

  useEffect(() => {
    fetchNotifications(filter);
  }, [filter]);

  // Real-time subscription for incoming notifications
  useEffect(() => {
    const wsScheme = window.location.protocol === 'https:' ? 'wss' : 'ws'
    const ws = new WebSocket(
      `${wsScheme}://${window.location.host}/api/v1/notifications/subscribe`
    )
    ws.onmessage = (event) => {
      try {
        const newNotif = JSON.parse(event.data) as Notification
        setNotifications((prev) => [newNotif, ...prev])
      } catch {
        // ignore invalid messages
      }
    }
    return () => {
      ws.close()
    }
  }, [])

  // Check existing push subscription
  useEffect(() => {
    if ('serviceWorker' in navigator && 'PushManager' in window) {
      navigator.serviceWorker.ready.then(registration =>
        registration.pushManager.getSubscription().then(sub => setPushSubscribed(!!sub))
      );
    }
  }, []);

  const markAsRead = async (notificationId?: string) => {
    try {
      if (notificationId) {
        await api.patch(`/notifications/${notificationId}`, { read: true });
        setNotifications(prev => 
          prev.map(n => n.id === notificationId ? { ...n, unread: false } : n)
        );
      } else {
        // Mark all as read
        await api.patch('/notifications', { read: true });
        setNotifications(prev => prev.map(n => ({ ...n, unread: false })));
      }
    } catch (err) {
      console.error('Failed to mark notification as read', err);
    }
  };

  const markSelectedAsRead = async () => {
    try {
      await Promise.all(
        selectedNotifications.map(id => api.patch(`/notifications/${id}`, { read: true }))
      );
      setNotifications(prev => 
        prev.map(n => selectedNotifications.includes(n.id) ? { ...n, unread: false } : n)
      );
      setSelectedNotifications([]);
    } catch (err) {
      console.error('Failed to mark notifications as read', err);
    }
  };

  const deleteNotification = async (notificationId: string) => {
    try {
      await api.delete(`/notifications/${notificationId}`);
      setNotifications(prev => prev.filter(n => n.id !== notificationId));
    } catch (err) {
      console.error('Failed to delete notification', err);
    }
  };

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case 'issue':
        return (
          <svg className="w-4 h-4 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.732L13.732 4.268c-.77-1.064-2.694-1.064-3.464 0L3.34 16.268C2.57 17.333 3.532 19 5.072 19z" />
          </svg>
        );
      case 'pull_request':
        return (
          <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
          </svg>
        );
      case 'mention':
        return (
          <svg className="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 12a4 4 0 10-8 0 4 4 0 008 0zm0 0v1.5a2.5 2.5 0 005 0V12a9 9 0 10-9 9m4.5-1.206a8.959 8.959 0 01-4.5 1.207" />
          </svg>
        );
      case 'security_alert':
        return (
          <svg className="w-4 h-4 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.732L13.732 4.268c-.77-1.064-2.694-1.064-3.464 0L3.34 16.268C2.57 17.333 3.532 19 5.072 19z" />
          </svg>
        );
      case 'repository_invite':
        return (
          <svg className="w-4 h-4 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
          </svg>
        );
      default:
        return (
          <svg className="w-4 h-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-5 5c-1.5-1.5-3.5-3.5-5-5z" />
          </svg>
        );
    }
  };

  const getReasonBadge = (reason: string) => {
    const reasonLabels = {
      subscribed: 'Subscribed',
      mentioned: 'Mentioned',
      assigned: 'Assigned',
      author: 'Author',
      comment: 'Comment',
      invitation: 'Invitation',
      security_alert: 'Security'
    };
    return reasonLabels[reason as keyof typeof reasonLabels] || reason;
  };

  const toggleNotificationSelection = (notificationId: string) => {
    setSelectedNotifications(prev => 
      prev.includes(notificationId)
        ? prev.filter(id => id !== notificationId)
        : [...prev, notificationId]
    );
  };

  const selectAllNotifications = () => {
    if (selectedNotifications.length === notifications.length) {
      setSelectedNotifications([]);
    } else {
      setSelectedNotifications(notifications.map(n => n.id));
    }
  };

  if (loading) {
    return (
      <AppLayout>
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse">
            <div className="h-8 bg-muted rounded w-1/3 mb-8"></div>
            <div className="space-y-4">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="h-20 bg-muted rounded"></div>
              ))}
            </div>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (error) {
    return (
      <AppLayout>
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <div className="text-red-600 text-lg mb-4">Error: {error}</div>
            <Button onClick={() => fetchNotifications()} disabled={loading}>
              {loading ? 'Retrying...' : 'Try Again'}
            </Button>
          </div>
        </div>
      </AppLayout>
    );
  }

  const unreadCount = Array.isArray(notifications) ? notifications.filter(n => n.unread).length : 0;

  return (
    <AppLayout>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-foreground">
              Notifications
              {unreadCount > 0 && (
                <Badge variant="default" className="ml-3">
                  {unreadCount} unread
                </Badge>
              )}
            </h1>
            <p className="text-muted-foreground mt-2">Stay updated on activity that matters to you</p>
          </div>
          
          <div className="flex items-center space-x-3">
            {selectedNotifications.length > 0 && (
              <Button variant="outline" size="sm" onClick={markSelectedAsRead} data-testid="mark-selected-as-read">
                Mark {selectedNotifications.length} as read
              </Button>
            )}
            {unreadCount > 0 && (
              <Button variant="outline" size="sm" onClick={() => markAsRead()} data-testid="mark-all-as-read">
                Mark all as read
              </Button>
            )}
            {isOnline && (
              pushSubscribed ? (
                <Button size="sm" onClick={async () => {
                  if (await unsubscribeFromPushNotifications()) {
                    setPushSubscribed(false);
                  }
                }}>
                  Disable Notifications
                </Button>
              ) : (
                <Button size="sm" onClick={async () => {
                  if (await subscribeToPushNotifications()) {
                    setPushSubscribed(true);
                  }
                }}>
                  Enable Notifications
                </Button>
              )
            )}
          </div>
        </div>

        {/* Filter Tabs */}
        <div className="border-b border-border mb-8">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setFilter('unread')}
              data-testid="filter-unread"
              className={`py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
                filter === 'unread'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
              }`}
            >
              <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 7.89a2 2 0 002.83 0L21 9M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
              </svg>
              Unread
              {unreadCount > 0 && (
                <Badge variant="secondary" className="ml-2">
                  {unreadCount}
                </Badge>
              )}
            </button>
            
            <button
              onClick={() => setFilter('all')}
              data-testid="filter-all"
              className={`py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
                filter === 'all'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
              }`}
            >
              <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 7a2 2 0 012-2h10a2 2 0 012 2v2M7 7h10" />
              </svg>
              All
              <Badge variant="secondary" className="ml-2">
                {notifications.length}
              </Badge>
            </button>
            
            <button
              onClick={() => setFilter('participating')}
              data-testid="filter-participating"
              className={`py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
                filter === 'participating'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
              }`}
            >
              <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8h2a2 2 0 012 2v6a2 2 0 01-2 2h-2v4l-4-4H9a1.994 1.994 0 01-1.414-.586m0 0L11 14h4a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2v4l.586-.586z" />
              </svg>
              Participating
            </button>
          </nav>
        </div>

        {/* Bulk Actions */}
        {notifications.length > 0 && (
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center space-x-4">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={selectedNotifications.length === notifications.length}
                  onChange={selectAllNotifications}
                  data-testid="select-all-notifications"
                  className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                />
                <span className="ml-2 text-sm text-foreground">
                  Select all {notifications.length} notifications
                </span>
              </label>
            </div>
            
            {selectedNotifications.length > 0 && (
              <div className="text-sm text-muted-foreground">
                {selectedNotifications.length} selected
              </div>
            )}
          </div>
        )}

        {/* Notifications List */}
        {notifications.length > 0 ? (
          <div className="space-y-1">
            {notifications.map((notification) => (
              <Card key={notification.id} className={notification.unread ? 'bg-blue-50 border-blue-200' : ''} data-testid="notification-item">
                <div className="p-4">
                  <div className="flex items-start space-x-4">
                    <input
                      type="checkbox"
                      checked={selectedNotifications.includes(notification.id)}
                      onChange={() => toggleNotificationSelection(notification.id)}
                      className="mt-1 h-4 w-4 text-primary focus:ring-primary border-border rounded"
                      aria-label={`Select notification: ${notification.subject.title}`}
                    />
                    
                    <div className="flex items-start space-x-3 flex-1">
                      {notification.unread && (
                        <div className="w-2 h-2 bg-blue-600 rounded-full mt-2"></div>
                      )}
                      
                      <div className="flex-shrink-0 mt-1" data-testid={`notification-icon-${notification.type}`}>
                        {getNotificationIcon(notification.type)}
                      </div>
                      
                      {notification.repository && (
                        <Avatar
                          src={notification.repository.owner.avatar_url}
                          alt={notification.repository.owner.username}
                          size="sm"
                          className="mt-1"
                        />
                      )}
                      
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center space-x-2 mb-1">
                          {notification.repository && (
                            <Link 
                              href={`/repositories/${notification.repository.full_name}`}
                              className="text-sm font-medium text-foreground hover:text-blue-600"
                            >
                              {notification.repository.full_name}
                            </Link>
                          )}
                          <Badge variant="outline" size="sm">
                            {getReasonBadge(notification.reason)}
                          </Badge>
                        </div>
                        
                        <div className="text-sm text-foreground mb-1">
                          {notification.subject.url ? (
                            <Link href={notification.subject.url} className="hover:text-blue-600">
                              {notification.subject.title}
                            </Link>
                          ) : (
                            notification.subject.title
                          )}
                        </div>
                        
                        <div className="text-xs text-muted-foreground">
                          {new Date(notification.updated_at).toLocaleString()}
                        </div>
                      </div>
                    </div>
                    
                    <div className="flex items-center space-x-2">
                      {notification.unread && (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => markAsRead(notification.id)}
                          data-testid={`mark-as-read-${notification.id}`}
                        >
                          Mark as read
                        </Button>
                      )}
                      
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => deleteNotification(notification.id)}
                        data-testid={`delete-notification-${notification.id}`}
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </Button>
                    </div>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        ) : (
          <Card>
            <div className="p-12 text-center">
              <svg className="w-16 h-16 mx-auto mb-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-5 5c-1.5-1.5-3.5-3.5-5-5z" />
              </svg>
                        <h3 className="text-lg font-medium text-foreground mb-2">All caught up!</h3>
          <p className="text-muted-foreground mb-4">
                {filter === 'unread' && "You have no unread notifications."}
                {filter === 'all' && "You have no notifications."}
                {filter === 'participating' && "You have no participating notifications."}
              </p>
              <Link href="/settings">
                <Button variant="outline">Manage notification settings</Button>
              </Link>
            </div>
          </Card>
        )}
      </div>
    </AppLayout>
  );
}
