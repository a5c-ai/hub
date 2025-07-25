'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Avatar } from '@/components/ui/Avatar';
import { useAuthStore } from '@/store/auth';
import api from '@/lib/api';

export default function ProfileEditPage() {
  const router = useRouter();
  const { user } = useAuthStore();
  const [saving, setSaving] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    username: '',
    bio: '',
    website: '',
    location: '',
    company: '',
    avatar_url: '',
    twitter: '',
    linkedin: ''
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
        company: '',
        avatar_url: user.avatar_url || '',
        twitter: '',
        linkedin: ''
      });
    }
  }, [user]);

  const handleSave = async () => {
    try {
      setSaving(true);
      await api.put('/user', formData);
      router.push('/settings');
    } catch (err) {
      console.error('Failed to save profile', err);
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    router.push('/settings');
  };

  return (
    <AppLayout>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
                  <h1 className="text-3xl font-bold text-foreground">Edit Profile</h1>
        <p className="text-muted-foreground mt-2">Update your personal information and profile details</p>
        </div>

        <Card>
          <div className="p-6">
            {/* Avatar Section */}
            <div className="mb-8">
              <h3 className="text-lg font-semibold text-foreground mb-4">Profile Picture</h3>
              <div className="flex items-center space-x-6">
                <Avatar
                  src={formData.avatar_url || user?.avatar_url}
                  alt={user?.username || 'User'}
                  size="2xl"
                />
                <div>
                  <Button variant="outline" size="sm" className="mr-3">
                    Upload New Picture
                  </Button>
                  <Button variant="outline" size="sm">
                    Remove Picture
                  </Button>
                              <p className="text-sm text-muted-foreground mt-2">
              Recommended: Square image, at least 400x400px
            </p>
                </div>
              </div>
            </div>

            {/* Basic Information */}
            <div className="mb-8">
              <h3 className="text-lg font-semibold text-foreground mb-4">Basic Information</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Full Name <span className="text-red-500">*</span>
                  </label>
                  <Input
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="Your full name"
                    required
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Username <span className="text-red-500">*</span>
                  </label>
                  <Input
                    value={formData.username}
                    onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                    placeholder="Your username"
                    required
                  />
                  <p className="text-sm text-muted-foreground mt-1">
                    This will be your unique identifier on the platform
                  </p>
                </div>
                
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Email Address <span className="text-red-500">*</span>
                  </label>
                  <Input
                    type="email"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    placeholder="your.email@example.com"
                    required
                  />
                </div>
              </div>
            </div>

            {/* Profile Details */}
            <div className="mb-8">
              <h3 className="text-lg font-semibold text-foreground mb-4">Profile Details</h3>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Bio
                  </label>
                  <textarea
                    value={formData.bio}
                    onChange={(e) => setFormData(prev => ({ ...prev, bio: e.target.value }))}
                    rows={3}
                    className="w-full px-3 py-2 border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent resize-none bg-background text-foreground placeholder:text-muted-foreground"
                    placeholder="Tell us about yourself..."
                  />
                  <p className="text-sm text-muted-foreground">
                    Maximum 160 characters
                  </p>
                  <p className="text-sm text-muted-foreground">
                    This will be displayed on your public profile
                  </p>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Company
                  </label>
                  <input
                    type="text"
                    value={formData.company}
                    onChange={(e) => setFormData(prev => ({ ...prev, company: e.target.value }))}
                    className="w-full px-3 py-2 border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent bg-background text-foreground placeholder:text-muted-foreground"
                    placeholder="Company name"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Location
                  </label>
                  <input
                    type="text"
                    value={formData.location}
                    onChange={(e) => setFormData(prev => ({ ...prev, location: e.target.value }))}
                    className="w-full px-3 py-2 border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent bg-background text-foreground placeholder:text-muted-foreground"
                    placeholder="City, Country"
                  />
                </div>
              </div>
            </div>

            {/* Social Links */}
            <div className="mb-8">
              <h3 className="text-lg font-semibold text-foreground mb-4">Social Links</h3>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Twitter Username
                  </label>
                  <div className="relative">
                    <span className="text-muted-foreground sm:text-sm">@</span>
                    <input
                      type="text"
                      value={formData.twitter}
                      onChange={(e) => setFormData(prev => ({ ...prev, twitter: e.target.value }))}
                      className="pl-8 w-full px-3 py-2 border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent bg-background text-foreground placeholder:text-muted-foreground"
                      placeholder="username"
                    />
                  </div>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    LinkedIn URL
                  </label>
                  <div className="relative">
                    <span className="text-muted-foreground sm:text-sm">linkedin.com/in/</span>
                    <input
                      type="text"
                      value={formData.linkedin}
                      onChange={(e) => setFormData(prev => ({ ...prev, linkedin: e.target.value }))}
                      className="pl-24 w-full px-3 py-2 border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent bg-background text-foreground placeholder:text-muted-foreground"
                      placeholder="username"
                    />
                  </div>
                </div>
              </div>
            </div>

            {/* Privacy Settings */}
            <div className="mb-8">
              <h3 className="text-lg font-semibold text-foreground mb-4">Privacy Settings</h3>
              <div className="space-y-4">
                <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                  <div>
                    <h4 className="font-medium text-foreground">Public Profile</h4>
                    <p className="text-sm text-muted-foreground">Make your profile visible to everyone</p>
                  </div>
                  <input 
                    type="checkbox" 
                    defaultChecked 
                    className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                  />
                </div>
                
                <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                  <div>
                    <h4 className="font-medium text-foreground">Show Email</h4>
                    <p className="text-sm text-muted-foreground">Display your email address on your profile</p>
                  </div>
                  <input 
                    type="checkbox" 
                    className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                  />
                </div>
                
                <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                  <div>
                    <h4 className="font-medium text-foreground">Activity Visibility</h4>
                    <p className="text-sm text-muted-foreground">Show your contributions and activity to others</p>
                  </div>
                  <input 
                    type="checkbox" 
                    defaultChecked 
                    className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                  />
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex items-center justify-end space-x-4 pt-6 border-t border-border">
              <Button variant="outline" onClick={handleCancel}>
                Cancel
              </Button>
              <Button onClick={handleSave} disabled={saving}>
                {saving ? 'Saving...' : 'Save Changes'}
              </Button>
            </div>
          </div>
        </Card>
      </div>
    </AppLayout>
  );
}