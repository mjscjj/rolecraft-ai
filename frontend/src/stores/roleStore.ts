import { create } from 'zustand';
import type { Role, CreateRoleRequest } from '../api/role';
import { roleApi } from '../api/role';

interface RoleState {
  roles: Role[];
  templates: Role[];
  currentRole: Role | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchRoles: (params?: { category?: string }) => Promise<void>;
  fetchTemplates: () => Promise<void>;
  fetchRole: (id: string) => Promise<void>;
  createRole: (data: CreateRoleRequest) => Promise<Role>;
  updateRole: (id: string, data: Partial<CreateRoleRequest>) => Promise<void>;
  deleteRole: (id: string) => Promise<void>;
  setCurrentRole: (role: Role | null) => void;
  clearError: () => void;
}

export const useRoleStore = create<RoleState>((set) => ({
  roles: [],
  templates: [],
  currentRole: null,
  isLoading: false,
  error: null,

  fetchRoles: async (params) => {
    set({ isLoading: true, error: null });
    try {
      const roles = await roleApi.list(params);
      set({ roles, isLoading: false });
    } catch (error: any) {
      set({
        error: error.message || 'Failed to fetch roles',
        isLoading: false,
      });
    }
  },

  fetchTemplates: async () => {
    set({ isLoading: true, error: null });
    try {
      const templates = await roleApi.getTemplates();
      set({ templates, isLoading: false });
    } catch (error: any) {
      set({
        error: error.message || 'Failed to fetch templates',
        isLoading: false,
      });
    }
  },

  fetchRole: async (id) => {
    set({ isLoading: true, error: null });
    try {
      const role = await roleApi.get(id);
      set({ currentRole: role, isLoading: false });
    } catch (error: any) {
      set({
        error: error.message || 'Failed to fetch role',
        isLoading: false,
      });
    }
  },

  createRole: async (data) => {
    set({ isLoading: true, error: null });
    try {
      const role = await roleApi.create(data);
      set((state) => ({
        roles: [role, ...state.roles],
        isLoading: false,
      }));
      return role;
    } catch (error: any) {
      set({
        error: error.message || 'Failed to create role',
        isLoading: false,
      });
      throw error;
    }
  },

  updateRole: async (id, data) => {
    set({ isLoading: true, error: null });
    try {
      const updatedRole = await roleApi.update(id, data);
      set((state) => ({
        roles: state.roles.map((r) => (r.id === id ? updatedRole : r)),
        currentRole: state.currentRole?.id === id ? updatedRole : state.currentRole,
        isLoading: false,
      }));
    } catch (error: any) {
      set({
        error: error.message || 'Failed to update role',
        isLoading: false,
      });
      throw error;
    }
  },

  deleteRole: async (id) => {
    set({ isLoading: true, error: null });
    try {
      await roleApi.delete(id);
      set((state) => ({
        roles: state.roles.filter((r) => r.id !== id),
        currentRole: state.currentRole?.id === id ? null : state.currentRole,
        isLoading: false,
      }));
    } catch (error: any) {
      set({
        error: error.message || 'Failed to delete role',
        isLoading: false,
      });
      throw error;
    }
  },

  setCurrentRole: (role) => set({ currentRole: role }),
  clearError: () => set({ error: null }),
}));

export default useRoleStore;
