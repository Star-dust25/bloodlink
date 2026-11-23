import { inject } from '@angular/core';
import { ActivatedRouteSnapshot, CanActivateFn, Router } from '@angular/router';
import { RolUsuario } from '../models/auth.model';
import { AuthService } from '../services/auth.service';

export const roleGuard: CanActivateFn = (route: ActivatedRouteSnapshot) => {
  const auth = inject(AuthService);
  const router = inject(Router);

  if (!auth.isAuthenticated()) {
    return router.createUrlTree(['/login']);
  }

  const allowedRoles = (route.data['roles'] as RolUsuario[] | undefined) ?? [];
  if (allowedRoles.length === 0 || auth.hasRole(allowedRoles)) {
    return true;
  }

  return router.createUrlTree([auth.defaultRouteForRole(auth.currentRole())]);
};
