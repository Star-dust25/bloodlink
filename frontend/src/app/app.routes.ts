import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';
import { roleGuard } from './core/guards/role.guard';
import { PublicLayoutComponent } from './layout/public-layout/public-layout.component';
import { WorkspaceLayoutComponent } from './layout/workspace-layout/workspace-layout.component';

export const routes: Routes = [
  // Ruta enfocada sin navbar del Home para el registro rápido de donantes (CP-05)
  {
    path: 'donantes/registro',
    loadComponent: () =>
      import('./features/donantes/registro-donante.component').then(
        (m) => m.RegistroDonanteComponent
      ),
  },

  // A. Rutas Públicas (dentro de PublicLayout)
  {
    path: '',
    component: PublicLayoutComponent,
    children: [
      {
        path: '',
        pathMatch: 'full',
        loadComponent: () =>
          import('./features/home/home.component').then((m) => m.HomeComponent),
      },
      {
        path: 'login',
        loadComponent: () =>
          import('./features/auth/login.component').then((m) => m.LoginComponent),
      },
      {
        path: 'interoperabilidad',
        loadComponent: () =>
          import('./features/interoperabilidad/interoperabilidad.component').then(
            (m) => m.InteroperabilidadComponent
          ),
      },
    ],
  },

  // B. Rutas del Workspace Institucional (protegidas por authGuard + roleGuard)
  {
    path: '',
    component: WorkspaceLayoutComponent,
    canActivate: [authGuard],
    children: [
      {
        path: 'dashboard',
        canActivate: [roleGuard],
        data: { roles: ['COORDINADOR', 'MEDICO', 'PERSONAL_BANCO'] },
        loadComponent: () =>
          import('./features/dashboard/dashboard.component').then(
            (m) => m.DashboardComponent
          ),
      },
      {
        path: 'emergencias',
        canActivate: [roleGuard],
        data: { roles: ['MEDICO', 'PERSONAL_BANCO', 'COORDINADOR'] },
        loadComponent: () =>
          import('./features/emergencias/emergencias-list.component').then(
            (m) => m.EmergenciasListComponent
          ),
      },
      {
        path: 'emergencias/nueva',
        canActivate: [roleGuard],
        data: { roles: ['MEDICO'] },
        loadComponent: () =>
          import('./features/emergencias/emergencia-nueva.component').then(
            (m) => m.EmergenciaNuevaComponent
          ),
      },
      {
        path: 'emergencias/:solicitudId',
        canActivate: [roleGuard],
        data: { roles: ['MEDICO', 'PERSONAL_BANCO', 'COORDINADOR'] },
        loadComponent: () =>
          import('./features/emergencias/emergencia-detalle.component').then(
            (m) => m.EmergenciaDetalleComponent
          ),
      },
      {
        path: 'inventario',
        canActivate: [roleGuard],
        data: { roles: ['PERSONAL_BANCO', 'COORDINADOR', 'MEDICO'] },
        loadComponent: () =>
          import('./features/inventario/inventario-list.component').then(
            (m) => m.InventarioListComponent
          ),
      },
      {
        path: 'inventario/nueva',
        canActivate: [roleGuard],
        data: { roles: ['PERSONAL_BANCO'] },
        loadComponent: () =>
          import('./features/inventario/unidad-nueva.component').then(
            (m) => m.UnidadNuevaComponent
          ),
      },
      {
        path: 'trazabilidad',
        canActivate: [roleGuard],
        data: { roles: ['COORDINADOR', 'PERSONAL_BANCO'] },
        loadComponent: () =>
          import('./features/trazabilidad/trazabilidad.component').then(
            (m) => m.TrazabilidadComponent
          ),
      },
      {
        path: 'trazabilidad/:unidadId',
        canActivate: [roleGuard],
        data: { roles: ['COORDINADOR', 'PERSONAL_BANCO'] },
        loadComponent: () =>
          import('./features/trazabilidad/trazabilidad.component').then(
            (m) => m.TrazabilidadComponent
          ),
      },
      {
        path: 'donantes',
        canActivate: [roleGuard],
        data: { roles: ['COORDINADOR', 'PERSONAL_BANCO'] },
        loadComponent: () =>
          import('./features/donantes/donantes-list.component').then(
            (m) => m.DonantesListComponent
          ),
      },
      {
        path: 'notificaciones',
        canActivate: [roleGuard],
        data: { roles: ['COORDINADOR', 'PERSONAL_BANCO', 'MEDICO'] },
        loadComponent: () =>
          import('./features/notificaciones/notificaciones.component').then(
            (m) => m.NotificacionesComponent
          ),
      },
    ],
  },

  {
    path: '**',
    redirectTo: '',
  },
];
