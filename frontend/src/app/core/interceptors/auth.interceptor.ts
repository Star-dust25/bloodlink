import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';
import { AuthService } from '../services/auth.service';

const PUBLIC_API_PATHS = [
  '/api/v1/auth/login',
  '/api/v1/donantes/validar-dni',
  '/api/v1/uddi/servicios',
  '/soap/v1/interoperabilidad',
  '/health',
];

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authService = inject(AuthService);
  const router = inject(Router);
  const token = authService.token();

  const isPublicPostDonante = req.url.endsWith('/api/v1/donantes') && req.method === 'POST';
  const isPublicEndpoint =
    isPublicPostDonante || PUBLIC_API_PATHS.some((path) => req.url.includes(path));

  const authReq =
    token && !isPublicEndpoint
      ? req.clone({
          setHeaders: {
            Authorization: `Bearer ${token}`,
          },
        })
      : req;

  return next(authReq).pipe(
    catchError((error: unknown) => {
      if (error instanceof HttpErrorResponse && error.status === 401 && !isPublicEndpoint) {
        authService.clearSession();
        void router.navigate(['/login']);
      }
      return throwError(() => error);
    })
  );
};
