import { Injectable, computed, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, tap } from 'rxjs';
import { API_CONFIG } from '../config/api.config';
import { LoginRequest, LoginResponse, RolUsuario, UsuarioDTO } from '../models/auth.model';

const TOKEN_STORAGE_KEY = 'bloodlink_jwt_token';
const USER_STORAGE_KEY = 'bloodlink_user_profile';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly router = inject(Router);

  private readonly tokenSignal = signal<string | null>(this.readTokenFromStorage());
  private readonly userSignal = signal<UsuarioDTO | null>(this.readUserFromStorage());

  readonly token = this.tokenSignal.asReadonly();
  readonly currentUser = this.userSignal.asReadonly();
  readonly isAuthenticated = computed(() => !!this.tokenSignal() && !!this.userSignal());
  readonly currentRole = computed<RolUsuario | null>(() => this.userSignal()?.rol ?? null);

  login(payload: LoginRequest): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(`${API_CONFIG.apiV1}/auth/login`, payload).pipe(
      tap((resp) => {
        this.persistSession(resp.accessToken, resp.usuario);
      })
    );
  }

  fetchProfile(): Observable<UsuarioDTO> {
    return this.http.get<UsuarioDTO>(`${API_CONFIG.apiV1}/auth/me`).pipe(
      tap((user) => {
        this.userSignal.set(user);
        this.safeSetStorage(USER_STORAGE_KEY, JSON.stringify(user));
      })
    );
  }

  hasRole(allowedRoles: readonly RolUsuario[]): boolean {
    const rol = this.currentRole();
    return !!rol && allowedRoles.includes(rol);
  }

  defaultRouteForRole(rol?: RolUsuario | null): string {
    switch (rol) {
      case 'MEDICO':
        return '/emergencias';
      case 'PERSONAL_BANCO':
        return '/inventario';
      case 'COORDINADOR':
        return '/dashboard';
      default:
        return '/login';
    }
  }

  clearSession(): void {
    this.tokenSignal.set(null);
    this.userSignal.set(null);
    this.safeRemoveStorage(TOKEN_STORAGE_KEY);
    this.safeRemoveStorage(USER_STORAGE_KEY);
  }

  logout(): void {
    this.clearSession();
    void this.router.navigate(['/login']);
  }

  private persistSession(token: string, user: UsuarioDTO): void {
    this.tokenSignal.set(token);
    this.userSignal.set(user);
    this.safeSetStorage(TOKEN_STORAGE_KEY, token);
    this.safeSetStorage(USER_STORAGE_KEY, JSON.stringify(user));
  }

  private readTokenFromStorage(): string | null {
    try {
      return typeof localStorage !== 'undefined' ? localStorage.getItem(TOKEN_STORAGE_KEY) : null;
    } catch {
      return null;
    }
  }

  private readUserFromStorage(): UsuarioDTO | null {
    try {
      if (typeof localStorage === 'undefined') return null;
      const raw = localStorage.getItem(USER_STORAGE_KEY);
      return raw ? (JSON.parse(raw) as UsuarioDTO) : null;
    } catch {
      return null;
    }
  }

  private safeSetStorage(key: string, value: string): void {
    try {
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem(key, value);
      }
    } catch {
      // ignore storage errors in restricted environments
    }
  }

  private safeRemoveStorage(key: string): void {
    try {
      if (typeof localStorage !== 'undefined') {
        localStorage.removeItem(key);
      }
    } catch {
      // ignore storage errors
    }
  }
}
