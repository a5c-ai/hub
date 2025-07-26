'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import api from '@/lib/api';

interface SAMLMapping {
  id: string;
  saml_group: string;
  organization_id: string;
  organization_name: string;
  role: 'admin' | 'member' | 'viewer';
  auto_create_org: boolean;
  attribute_name: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface SAMLProvider {
  id: string;
  name: string;
  entity_id: string;
  sso_url: string;
  certificate: string;
  is_active: boolean;
}

interface SAMLAttributeMapping {
  [key: string]: string;
}

interface Organization {
  id: string;
  name: string;
  display_name: string;
}

export function SAMLOrganizationMapper() {
  const [mappings, setMappings] = useState<SAMLMapping[]>([]);
  const [providers, setProviders] = useState<SAMLProvider[]>([]);
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [attributeMapping, setAttributeMapping] = useState<SAMLAttributeMapping>({});
  
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [showAddModal, setShowAddModal] = useState(false);
  const [showAttributeModal, setShowAttributeModal] = useState(false);
  
  // New mapping form state
  const [newMapping, setNewMapping] = useState({
    saml_group: '',
    organization_id: '',
    role: 'member' as 'admin' | 'member' | 'viewer',
    attribute_name: 'groups',
    auto_create_org: false
  });

  // New attribute mapping state
  const [newAttribute, setNewAttribute] = useState({
    saml_attribute: '',
    mapped_field: ''
  });

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [mappingsResponse, providersResponse, orgsResponse, attributesResponse] = await Promise.all([
        api.get('/admin/auth/saml/org-mapping'),
        api.get('/admin/auth/saml/providers'),
        api.get('/admin/organizations'),
        api.get('/admin/auth/saml/attribute-mapping')
      ]);

      setMappings(mappingsResponse.data.mappings || []);
      setProviders(providersResponse.data.providers || []);
      setOrganizations(orgsResponse.data.organizations || []);
      setAttributeMapping(attributesResponse.data.mapping || {});
    } catch (error) {
      console.error('Failed to fetch SAML data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleAddMapping = async () => {
    if (!newMapping.saml_group.trim() || (!newMapping.organization_id && !newMapping.auto_create_org)) {
      return;
    }

    try {
      const response = await api.post('/admin/auth/saml/org-mapping', newMapping);
      setMappings(prev => [...prev, response.data]);
      setNewMapping({
        saml_group: '',
        organization_id: '',
        role: 'member',
        attribute_name: 'groups',
        auto_create_org: false
      });
      setShowAddModal(false);
    } catch (error) {
      console.error('Failed to add SAML mapping:', error);
    }
  };

  const handleDeleteMapping = async (mappingId: string) => {
    if (!confirm('Are you sure you want to delete this SAML mapping?')) {
      return;
    }

    try {
      await api.delete(`/admin/auth/saml/org-mapping/${mappingId}`);
      setMappings(prev => prev.filter(m => m.id !== mappingId));
    } catch (error) {
      console.error('Failed to delete SAML mapping:', error);
    }
  };

  const handleToggleMapping = async (mappingId: string, isActive: boolean) => {
    try {
      await api.patch(`/admin/auth/saml/org-mapping/${mappingId}`, { is_active: !isActive });
      setMappings(prev => prev.map(m => 
        m.id === mappingId ? { ...m, is_active: !isActive } : m
      ));
    } catch (error) {
      console.error('Failed to toggle SAML mapping:', error);
    }
  };

  const handleUpdateAttributeMapping = async () => {
    setSaving(true);
    try {
      await api.put('/admin/auth/saml/attribute-mapping', { mapping: attributeMapping });
    } catch (error) {
      console.error('Failed to update attribute mapping:', error);
    } finally {
      setSaving(false);
    }
  };

  const handleAddAttribute = () => {
    if (!newAttribute.saml_attribute.trim() || !newAttribute.mapped_field.trim()) {
      return;
    }

    setAttributeMapping(prev => ({
      ...prev,
      [newAttribute.saml_attribute]: newAttribute.mapped_field
    }));

    setNewAttribute({
      saml_attribute: '',
      mapped_field: ''
    });
  };

  const handleRemoveAttribute = (attributeName: string) => {
    setAttributeMapping(prev => {
      const newMapping = { ...prev };
      delete newMapping[attributeName];
      return newMapping;
    });
  };

  const handleTestMapping = async (mappingId: string) => {
    try {
      const response = await api.post(`/admin/auth/saml/org-mapping/${mappingId}/test`);
      alert(`Test result: ${response.data.message}`);
    } catch (error) {
      console.error('Failed to test SAML mapping:', error);
      alert('Test failed. Check console for details.');
    }
  };

  const filteredMappings = mappings.filter(mapping =>
    mapping.saml_group.toLowerCase().includes(searchQuery.toLowerCase()) ||
    mapping.organization_name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getRoleBadgeColor = (role: string) => {
    switch (role) {
      case 'admin': return 'bg-red-100 text-red-800';
      case 'member': return 'bg-blue-100 text-blue-800';
      case 'viewer': return 'bg-green-100 text-green-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const commonSAMLAttributes = [
    'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress',
    'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name',
    'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname',
    'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname',
    'http://schemas.microsoft.com/ws/2008/06/identity/claims/groups',
    'groups',
    'roles',
    'department',
    'title'
  ];

  const mappableFields = ['email', 'username', 'name', 'groups', 'roles', 'department'];

  return (
    <div className="space-y-6">
      {/* SAML Providers Status */}
      <Card>
        <div className="p-6">
          <h3 className="text-lg font-semibold text-foreground mb-4">SAML Providers</h3>
          
          {providers.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No SAML providers configured
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4">
              {providers.map((provider) => (
                <div key={provider.id} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium text-foreground">{provider.name}</h4>
                      <p className="text-sm text-muted-foreground">Entity ID: {provider.entity_id}</p>
                      <p className="text-sm text-muted-foreground">SSO URL: {provider.sso_url}</p>
                    </div>
                    <Badge className={provider.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}>
                      {provider.is_active ? 'Active' : 'Inactive'}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </Card>

      {/* SAML Attribute Mapping */}
      <Card>
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-semibold text-foreground">SAML Attribute Mapping</h3>
            <div className="flex space-x-2">
              <Button
                variant="outline"
                onClick={() => setShowAttributeModal(true)}
              >
                Manage Attributes
              </Button>
              <Button onClick={handleUpdateAttributeMapping} disabled={saving}>
                {saving ? 'Saving...' : 'Save Mapping'}
              </Button>
            </div>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {Object.entries(attributeMapping).map(([samlAttr, mappedField]) => (
              <div key={samlAttr} className="border rounded-lg p-3">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="font-medium text-sm">{samlAttr}</div>
                    <div className="text-muted-foreground text-xs">→ {mappedField}</div>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleRemoveAttribute(samlAttr)}
                    className="text-red-600 hover:text-red-700"
                  >
                    ×
                  </Button>
                </div>
              </div>
            ))}
          </div>
          
          {Object.keys(attributeMapping).length === 0 && (
            <div className="text-center py-8 text-muted-foreground">
              No attribute mappings configured
            </div>
          )}
        </div>
      </Card>

      {/* Organization Mapping */}
      <Card>
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-semibold text-foreground">SAML Group to Organization Mapping</h3>
            <Button onClick={() => setShowAddModal(true)}>
              Add Mapping
            </Button>
          </div>

          <div className="mb-4">
            <Input
              placeholder="Search mappings by SAML group or organization..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="max-w-md"
            />
          </div>

          {loading ? (
            <div className="text-center py-8">Loading SAML mappings...</div>
          ) : filteredMappings.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No SAML group mappings found
            </div>
          ) : (
            <div className="space-y-3">
              {filteredMappings.map((mapping) => (
                <div key={mapping.id} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2">
                        <span className="font-medium text-foreground">{mapping.saml_group}</span>
                        <span className="text-muted-foreground">→</span>
                        <span className="text-foreground">{mapping.organization_name}</span>
                        <Badge className={getRoleBadgeColor(mapping.role)}>
                          {mapping.role}
                        </Badge>
                        {mapping.auto_create_org && (
                          <Badge variant="outline">Auto-create</Badge>
                        )}
                        <Badge className={mapping.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}>
                          {mapping.is_active ? 'Active' : 'Inactive'}
                        </Badge>
                      </div>
                      <div className="text-sm text-muted-foreground mt-1">
                        Attribute: {mapping.attribute_name} • Created: {new Date(mapping.created_at).toLocaleDateString()}
                        {mapping.updated_at !== mapping.created_at && (
                          <span> • Updated: {new Date(mapping.updated_at).toLocaleDateString()}</span>
                        )}
                      </div>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleTestMapping(mapping.id)}
                      >
                        Test
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleToggleMapping(mapping.id, mapping.is_active)}
                        className={mapping.is_active ? 'text-orange-600' : 'text-green-600'}
                      >
                        {mapping.is_active ? 'Disable' : 'Enable'}
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleDeleteMapping(mapping.id)}
                        className="text-red-600 hover:text-red-700"
                      >
                        Delete
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </Card>

      {/* Add Mapping Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <Card className="w-full max-w-md">
            <div className="p-6">
              <h3 className="text-lg font-semibold text-foreground mb-4">Add SAML Group Mapping</h3>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    SAML Group Name
                  </label>
                  <Input
                    value={newMapping.saml_group}
                    onChange={(e) => setNewMapping(prev => ({ ...prev, saml_group: e.target.value }))}
                    placeholder="e.g., developers, admins, viewers"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    SAML Attribute Name
                  </label>
                  <select
                    value={newMapping.attribute_name}
                    onChange={(e) => setNewMapping(prev => ({ ...prev, attribute_name: e.target.value }))}
                    className="w-full px-3 py-2 border rounded-md"
                  >
                    {commonSAMLAttributes.map((attr) => (
                      <option key={attr} value={attr}>
                        {attr}
                      </option>
                    ))}
                  </select>
                </div>
                
                <div className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    id="auto-create"
                    checked={newMapping.auto_create_org}
                    onChange={(e) => setNewMapping(prev => ({ ...prev, auto_create_org: e.target.checked }))}
                    className="rounded"
                  />
                  <label htmlFor="auto-create" className="text-sm font-medium">
                    Auto-create organization if it does not exist
                  </label>
                </div>

                {!newMapping.auto_create_org && (
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">
                      Target Organization
                    </label>
                    <select
                      value={newMapping.organization_id}
                      onChange={(e) => setNewMapping(prev => ({ ...prev, organization_id: e.target.value }))}
                      className="w-full px-3 py-2 border rounded-md"
                    >
                      <option value="">Select an organization</option>
                      {organizations.map((org) => (
                        <option key={org.id} value={org.id}>
                          {org.display_name || org.name}
                        </option>
                      ))}
                    </select>
                  </div>
                )}
                
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Default Role
                  </label>
                  <select
                    value={newMapping.role}
                    onChange={(e) => setNewMapping(prev => ({ ...prev, role: e.target.value as 'admin' | 'member' | 'viewer' }))}
                    className="w-full px-3 py-2 border rounded-md"
                  >
                    <option value="viewer">Viewer</option>
                    <option value="member">Member</option>
                    <option value="admin">Admin</option>
                  </select>
                </div>
              </div>
              
              <div className="flex justify-end space-x-2 mt-6">
                <Button
                  variant="outline"
                  onClick={() => setShowAddModal(false)}
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleAddMapping}
                  disabled={!newMapping.saml_group.trim() || (!newMapping.organization_id && !newMapping.auto_create_org)}
                >
                  Add Mapping
                </Button>
              </div>
            </div>
          </Card>
        </div>
      )}

      {/* Attribute Mapping Modal */}
      {showAttributeModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <Card className="w-full max-w-lg">
            <div className="p-6">
              <h3 className="text-lg font-semibold text-foreground mb-4">Manage SAML Attribute Mapping</h3>
              
              <div className="space-y-4 mb-4">
                <div className="flex space-x-2">
                  <div className="flex-1">
                    <select
                      value={newAttribute.saml_attribute}
                      onChange={(e) => setNewAttribute(prev => ({ ...prev, saml_attribute: e.target.value }))}
                      className="w-full px-3 py-2 border rounded-md"
                    >
                      <option value="">Select SAML attribute</option>
                      {commonSAMLAttributes.map((attr) => (
                        <option key={attr} value={attr}>
                          {attr}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="flex-1">
                    <select
                      value={newAttribute.mapped_field}
                      onChange={(e) => setNewAttribute(prev => ({ ...prev, mapped_field: e.target.value }))}
                      className="w-full px-3 py-2 border rounded-md"
                    >
                      <option value="">Select mapped field</option>
                      {mappableFields.map((field) => (
                        <option key={field} value={field}>
                          {field}
                        </option>
                      ))}
                    </select>
                  </div>
                  <Button
                    onClick={handleAddAttribute}
                    disabled={!newAttribute.saml_attribute || !newAttribute.mapped_field}
                  >
                    Add
                  </Button>
                </div>
              </div>

              <div className="border rounded-lg p-3 max-h-64 overflow-y-auto">
                {Object.entries(attributeMapping).map(([samlAttr, mappedField]) => (
                  <div key={samlAttr} className="flex items-center justify-between py-2 border-b last:border-b-0">
                    <div className="flex-1 text-sm">
                      <span className="font-medium">{samlAttr}</span>
                      <span className="text-muted-foreground"> → {mappedField}</span>
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleRemoveAttribute(samlAttr)}
                      className="text-red-600 hover:text-red-700"
                    >
                      Remove
                    </Button>
                  </div>
                ))}
              </div>
              
              <div className="flex justify-end space-x-2 mt-6">
                <Button
                  variant="outline"
                  onClick={() => setShowAttributeModal(false)}
                >
                  Close
                </Button>
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
}