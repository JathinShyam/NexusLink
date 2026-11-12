import { createContext } from "react";
import type { AuthUser } from "../api/client";

export type AuthStatus = "loading" | "anonymous" | "authenticated";
export type AuthResult = { ok: true; user: AuthUser } | { ok: false };

export type AuthContextValue = {
  status: AuthStatus;
  user: AuthUser | null;
  token: string;
  pending: boolean;
  error: string | null;
  notice: "signed-out" | null;
  isAuthenticated: boolean;
  signIn: (email: string, password: string) => Promise<AuthResult>;
  register: (email: string, password: string, displayName: string) => Promise<AuthResult>;
  signOut: () => Promise<void>;
  clearSession: () => void;
  clearError: () => void;
};

export const AuthContext = createContext<AuthContextValue | null>(null);
