'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Avatar } from '@/components/ui/Avatar';
import { ThemeToggle } from '@/components/ui/ThemeToggle';
import { useAuthStore } from '@/store/auth';
import api from '@/lib/api';

export default function SettingsPage() {
  const { user } = useAuthStore();
  const [activeTab, setActiveTab] = useState('profile');
  const [saving, setSaving] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    username: '',
    bio: '',
    website: '',
    location: '',
    company: ''
  });

  useEffect(() => {
    if (user) {
      setFormData({
        name: user.name || '',
        email: user.email || '',
        username: user.username || '',
        bio: '',
        website: '',
        location: '',
        company: ''
      });
    }
  }, [user]);

  const handleSave = async () => {
    try {
      setSaving(true);
      await api.put('/user', formData);
      // Show success message
    } catch (err) {
      console.error('Failed to save settings', err);
    } finally {
      setSaving(false);
    }
  };

  const tabs = [
    { id: 'profile', name: 'Profile', icon: '👤' },
    { id: 'account', name: 'Account', icon: '⚙️' },
    { id: 'security', name: 'Security', icon: '🔒' },
    { id: 'notifications', name: 'Notifications', icon: '🔔' },
    { id: 'billing', name: 'Billing', icon: '💳' },
  ];

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground">Settings</h1>
          <p className="text-muted-foreground mt-2">Manage your account settings and preferences</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          {/* Sidebar */}
          <div className="lg:col-span-1">
            <nav className="space-y-1">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`w-full flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                    activeTab === tab.id
                      ? 'bg-primary/10 text-primary border-r-2 border-primary'
                      : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                  }`}
                >
                  <span className="mr-3">{tab.icon}</span>
                  {tab.name}
                </button>
              ))}
            </nav>
          </div>

          {/* Main Content */}
          <div className="lg:col-span-3">
            {activeTab === 'profile' && (
              <div className="space-y-6">
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Profile Information</h3>
                    
                    {/* Avatar */}
                    <div className="mb-6">
                      <label className="block text-sm font-medium text-foreground mb-2">
                        Profile Picture
                      </label>
                      <div className="flex items-center space-x-4">
                        <Avatar
                          src={user?.avatar_url}
                          alt={user?.username || 'User'}
                          size="xl"
                        />
                        <Button variant="outline" size="sm">
                          Change Avatar
                        </Button>
                      </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Full Name
                        </label>
                        <Input
                          value={formData.name}
                          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                          placeholder="Your full name"
                        />
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Username
                        </label>
                        <Input
                          value={formData.username}
                          onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                          placeholder="Your username"
                        />
                      </div>
                      
                      <div className="md:col-span-2">
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Email
                        </label>
                        <Input
                          type="email"
                          value={formData.email}
                          onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                          placeholder="your.email@example.com"
                        />
                      </div>
                      
                      <div className="md:col-span-2">
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Bio
                        </label>
                        <textarea
                          value={formData.bio}
                          onChange={(e) => setFormData({ ...formData, bio: e.target.value })}
                          placeholder="Tell us about yourself..."
                          rows={3}
                          className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent disabled:cursor-not-allowed disabled:opacity-50"
                        />
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Website
                        </label>
                        <Input
                          value={formData.website}
                          onChange={(e) => setFormData({ ...formData, website: e.target.value })}
                          placeholder="https://yourwebsite.com"
                        />
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Location
                        </label>
                        <Input
                          value={formData.location}
                          onChange={(e) => setFormData({ ...formData, location: e.target.value })}
                          placeholder="City, Country"
                        />
                      </div>
                      
                      <div className="md:col-span-2">
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Company
                        </label>
                        <Input
                          value={formData.company}
                          onChange={(e) => setFormData({ ...formData, company: e.target.value })}
                          placeholder="Your company or organization"
                        />
                      </div>
                    </div>
                    
                    <div className="mt-6">
                      <Button onClick={handleSave} disabled={saving}>
                        {saving ? 'Saving...' : 'Save Changes'}
                      </Button>
                    </div>
                  </div>
                </Card>
              </div>
            )}

            {activeTab === 'account' && (
              <div className="space-y-6">
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Account Settings</h3>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                        <div>
                          <h4 className="font-medium text-foreground">Account Type</h4>
                          <p className="text-sm text-muted-foreground">Free account with basic features</p>
                        </div>
                        <Button size="sm" variant="outline">Upgrade</Button>
                      </div>
                      
                      <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                        <div>
                          <h4 className="font-medium text-foreground">Account Status</h4>
                          <p className="text-sm text-muted-foreground">Active since {new Date(user?.created_at || '').toLocaleDateString()}</p>
                        </div>
                        <span className="px-2 py-1 text-xs font-medium bg-success text-success-foreground rounded-full">
                          Active
                        </span>
                      </div>
                    </div>
                  </div>
                </Card>

                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Appearance</h3>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                        <div>
                          <h4 className="font-medium text-foreground">Theme</h4>
                          <p className="text-sm text-muted-foreground">Choose your preferred color scheme</p>
                        </div>
                        <ThemeToggle />
                      </div>
                    </div>
                  </div>
                </Card>
                
                <Card className="border-destructive/20">
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-destructive mb-4">Danger Zone</h3>
                    <div className="space-y-4">
                      <div className="border border-destructive/20 rounded-lg p-4">
                        <div className="flex items-center justify-between">
                          <div>
                            <h4 className="font-medium text-destructive">Delete Account</h4>
                            <p className="text-sm text-destructive/80">Permanently delete your account and all associated data</p>
                          </div>
                          <Button variant="outline" size="sm" className="border-destructive/30 text-destructive hover:bg-destructive/10">
                            Delete Account
                          </Button>
                        </div>
                      </div>
                    </div>
                  </div>
                </Card>
              </div>
            )}

            {activeTab === 'security' && (
              <div className="space-y-6">
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Password & Authentication</h3>
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Current Password
                        </label>
                        <Input type="password" placeholder="Enter current password" />
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          New Password
                        </label>
                        <Input type="password" placeholder="Enter new password" />
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Confirm New Password
                        </label>
                        <Input type="password" placeholder="Confirm new password" />
                      </div>
                      
                      <Button>Change Password</Button>
                    </div>
                  </div>
                </Card>
                
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Two-Factor Authentication</h3>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                        <div>
                          <h4 className="font-medium text-foreground">Authenticator App</h4>
                          <p className="text-sm text-muted-foreground">Use an authenticator app to generate verification codes</p>
                        </div>
                        <Button size="sm">Enable</Button>
                      </div>
                      
                      <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                        <div>
                          <h4 className="font-medium text-foreground">SMS Authentication</h4>
                          <p className="text-sm text-muted-foreground">Receive verification codes via text message</p>
                        </div>
                        <Button size="sm" variant="outline">Enable</Button>
                      </div>
                    </div>
                  </div>
                </Card>
              </div>
            )}

            {activeTab === 'notifications' && (
              <div className="space-y-6">
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Email Notifications</h3>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <div>
                          <h4 className="font-medium text-foreground">Issues and Pull Requests</h4>
                          <p className="text-sm text-muted-foreground">Notifications about issues and pull requests you&apos;re involved in</p>

                        </div>
                        <input type="checkbox" defaultChecked className="h-4 w-4 text-primary focus:ring-primary border-border rounded" />
                      </div>
                      
                      <div className="flex items-center justify-between">
                        <div>
                          <h4 className="font-medium text-foreground">Repository Updates</h4>
                          <p className="text-sm text-muted-foreground">Notifications about repositories you&apos;re watching</p>
                        </div>
                        <input type="checkbox" defaultChecked className="h-4 w-4 text-primary focus:ring-primary border-border rounded" />
                      </div>
                      
                      <div className="flex items-center justify-between">
                        <div>
                          <h4 className="font-medium text-foreground">Security Alerts</h4>
                          <p className="text-sm text-muted-foreground">Notifications about security vulnerabilities</p>
                        </div>
                        <input type="checkbox" defaultChecked className="h-4 w-4 text-primary focus:ring-primary border-border rounded" />
                      </div>
                    </div>
                  </div>
                </Card>
                
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Web Notifications</h3>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <div>
                          <h4 className="font-medium text-foreground">Browser Notifications</h4>
                          <p className="text-sm text-muted-foreground">Show notifications in your browser</p>
                        </div>
                        <Button size="sm" variant="outline">Enable</Button>
                      </div>
                    </div>
                  </div>
                </Card>
              </div>
            )}

            {activeTab === 'billing' && (
              <div className="space-y-6">
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Billing Information</h3>
                    <div className="space-y-4">
                      <div className="p-4 bg-muted rounded-lg">
                        <div className="flex items-center justify-between mb-2">
                          <h4 className="font-medium text-foreground">Current Plan</h4>
                          <span className="px-2 py-1 text-xs font-medium bg-primary text-primary-foreground rounded-full">
                            Free
                          </span>
                        </div>
                        <p className="text-sm text-muted-foreground">
                          Free plan with unlimited public repositories
                        </p>
                      </div>
                      
                      <div className="border-t pt-4">
                        <h4 className="font-medium text-foreground mb-3">Upgrade Options</h4>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                          <div className="border border-border rounded-lg p-4 bg-card">
                            <h5 className="font-medium text-foreground mb-2">Pro</h5>
                            <div className="text-2xl font-bold text-foreground mb-2">$4<span className="text-sm font-normal text-muted-foreground">/month</span></div>
                            <ul className="text-sm text-muted-foreground space-y-1 mb-4">
                              <li>• Unlimited private repositories</li>
                              <li>• Advanced collaboration features</li>
                              <li>• Priority support</li>
                            </ul>
                            <Button size="sm" className="w-full">Upgrade to Pro</Button>
                          </div>
                          
                          <div className="border border-border rounded-lg p-4 bg-card">
                            <h5 className="font-medium text-foreground mb-2">Team</h5>
                            <div className="text-2xl font-bold text-foreground mb-2">$4<span className="text-sm font-normal text-muted-foreground">/user/month</span></div>
                            <ul className="text-sm text-muted-foreground space-y-1 mb-4">
                              <li>• Everything in Pro</li>
                              <li>• Team management</li>
                              <li>• Advanced security features</li>
                            </ul>
                            <Button size="sm" variant="outline" className="w-full">Upgrade to Team</Button>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </Card>
              </div>
            )}
          </div>
        </div>
      </div>
    </AppLayout>
  );
}