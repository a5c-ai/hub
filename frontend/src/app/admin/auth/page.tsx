'use client';

import React, { useState } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { SAMLConfiguration } from '@/components/auth/SAMLConfiguration';
import { LDAPConfiguration } from '@/components/auth/LDAPConfiguration';
import { SessionManagement } from '@/components/auth/SessionManagement';
import { SecurityDashboard } from '@/components/auth/SecurityDashboard';

type TabType = 'overview' | 'saml' | 'ldap' | 'sessions' | 'security';

export default function AdminAuthPage() {
  const [activeTab, setActiveTab] = useState<TabType>('overview');

  const tabs = [
    { id: 'overview' as TabType, name: 'Overview', icon: '📊' },
    { id: 'saml' as TabType, name: 'SAML/SSO', icon: '🔐' },
    { id: 'ldap' as TabType, name: 'LDAP/AD', icon: '🏢' },
    { id: 'sessions' as TabType, name: 'Sessions', icon: '💻' },
    { id: 'security' as TabType, name: 'Security', icon: '🛡️' },
  ];

  const renderTabContent = () => {
    switch (activeTab) {
      case 'overview':
        return (
          <div className="space-y-6">
            <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold text-foreground mb-4">Enterprise Authentication Overview</h3>
                <p className="text-muted-foreground mb-6">
                  Configure and manage enterprise-grade authentication features for your organization.
                </p>
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="p-4 border rounded-lg">
                    <div className="flex items-center space-x-3 mb-3">
                      <span className="text-2xl">🔐</span>
                      <h4 className="font-semibold text-foreground">SAML Single Sign-On</h4>
                    </div>
                    <p className="text-sm text-muted-foreground mb-3">
                      Configure SAML 2.0 identity providers for seamless single sign-on experience.
                    </p>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setActiveTab('saml')}
                    >
                      Configure SAML
                    </Button>
                  </div>

                  <div className="p-4 border rounded-lg">
                    <div className="flex items-center space-x-3 mb-3">
                      <span className="text-2xl">🏢</span>
                      <h4 className="font-semibold text-foreground">LDAP/Active Directory</h4>
                    </div>
                    <p className="text-sm text-muted-foreground mb-3">
                      Connect to your organization&apos;s LDAP or Active Directory for user authentication.
                    </p>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setActiveTab('ldap')}
                    >
                      Configure LDAP
                    </Button>
                  </div>

                  <div className="p-4 border rounded-lg">
                    <div className="flex items-center space-x-3 mb-3">
                      <span className="text-2xl">💻</span>
                      <h4 className="font-semibold text-foreground">Session Management</h4>
                    </div>
                    <p className="text-sm text-muted-foreground mb-3">
                      Monitor and control user sessions, including device tracking and session timeouts.
                    </p>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setActiveTab('sessions')}
                    >
                      Manage Sessions
                    </Button>
                  </div>

                  <div className="p-4 border rounded-lg">
                    <div className="flex items-center space-x-3 mb-3">
                      <span className="text-2xl">🛡️</span>
                      <h4 className="font-semibold text-foreground">Security Monitoring</h4>
                    </div>
                    <p className="text-sm text-muted-foreground mb-3">
                      Monitor authentication events, configure rate limiting, and track security incidents.
                    </p>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setActiveTab('security')}
                    >
                      View Security Dashboard
                    </Button>
                  </div>
                </div>
              </div>
            </Card>

            <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold text-foreground mb-4">Quick Stats</h3>
                <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                  <div className="text-center">
                    <div className="text-2xl font-bold text-blue-600">2</div>
                    <div className="text-sm text-muted-foreground">SAML Providers</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-green-600">1</div>
                    <div className="text-sm text-muted-foreground">LDAP Connections</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-purple-600">247</div>
                    <div className="text-sm text-muted-foreground">Active Sessions</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-orange-600">3</div>
                    <div className="text-sm text-muted-foreground">Security Alerts</div>
                  </div>
                </div>
              </div>
            </Card>
          </div>
        );
      
      case 'saml':
        return (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-semibold text-foreground mb-2">SAML Configuration</h3>
              <p className="text-muted-foreground">
                Configure SAML 2.0 Single Sign-On providers for your organization.
              </p>
            </div>
            <SAMLConfiguration />
          </div>
        );
      
      case 'ldap':
        return (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-semibold text-foreground mb-2">LDAP Configuration</h3>
              <p className="text-muted-foreground">
                Configure LDAP or Active Directory authentication for your organization.
              </p>
            </div>
            <LDAPConfiguration />
          </div>
        );
      
      case 'sessions':
        return (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-semibold text-foreground mb-2">Session Management</h3>
              <p className="text-muted-foreground">
                Monitor and manage user sessions across your organization.
              </p>
            </div>
            <SessionManagement isAdminView={true} />
          </div>
        );
      
      case 'security':
        return (
          <div className="space-y-6">
            <SecurityDashboard />
          </div>
        );
      
      default:
        return null;
    }
  };

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground">Enterprise Authentication</h1>
          <p className="text-muted-foreground mt-2">Configure and manage authentication settings for your organization</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-5 gap-8">
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
          <div className="lg:col-span-4">
            {renderTabContent()}
          </div>
        </div>
      </div>
    </AppLayout>
  );
}