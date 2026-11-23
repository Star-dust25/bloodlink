import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { RolUsuario } from '../../core/models/auth.model';
import { AuthService } from '../../core/services/auth.service';
import { formatHospital, formatRol } from '../../core/utils/label.util';

interface NavItem {
  label: string;
  route: string;
  icon: 'dashboard' | 'alert' | 'plus' | 'droplet' | 'box' | 'history' | 'users' | 'bell';
  roles: readonly RolUsuario[];
}

const WORKSPACE_NAV_ITEMS: readonly NavItem[] = [
  {
    label: 'Resumen general',
    route: '/dashboard',
    icon: 'dashboard',
    roles: ['COORDINADOR', 'MEDICO', 'PERSONAL_BANCO'],
  },
  {
    label: 'Emergencias',
    route: '/emergencias',
    icon: 'alert',
    roles: ['MEDICO', 'PERSONAL_BANCO', 'COORDINADOR'],
  },
  {
    label: 'Pedir unidades',
    route: '/emergencias/nueva',
    icon: 'plus',
    roles: ['MEDICO'],
  },
  {
    label: 'Inventario de sangre',
    route: '/inventario',
    icon: 'droplet',
    roles: ['PERSONAL_BANCO', 'COORDINADOR', 'MEDICO'],
  },
  {
    label: 'Ingresar unidad',
    route: '/inventario/nueva',
    icon: 'box',
    roles: ['PERSONAL_BANCO'],
  },
  {
    label: 'Historial de unidad',
    route: '/trazabilidad',
    icon: 'history',
    roles: ['COORDINADOR', 'PERSONAL_BANCO'],
  },
  {
    label: 'Donantes voluntarios',
    route: '/donantes',
    icon: 'users',
    roles: ['COORDINADOR', 'PERSONAL_BANCO'],
  },
  {
    label: 'Alertas y avisos',
    route: '/notificaciones',
    icon: 'bell',
    roles: ['COORDINADOR', 'PERSONAL_BANCO', 'MEDICO'],
  },
];

@Component({
  selector: 'app-workspace-layout',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="workspace-shell">
      <aside class="sidebar no-print" [class.open]="mobileMenuOpen()">
        <div class="sidebar-top">
          <a routerLink="/" class="sidebar-brand">BloodLink</a>
          <button
            type="button"
            class="mobile-toggle"
            (click)="toggleMobileMenu()"
            aria-label="Alternar menú"
          >
            {{ mobileMenuOpen() ? 'Cerrar' : 'Menú' }}
          </button>
        </div>

        <nav class="sidebar-nav" aria-label="Navegación institucional">
          @for (item of visibleNavItems(); track item.route) {
            <a
              [routerLink]="item.route"
              routerLinkActive="active"
              [routerLinkActiveOptions]="{ exact: item.route === '/emergencias' || item.route === '/inventario' }"
              (click)="closeMobileMenu()"
            >
              <span class="nav-dot" aria-hidden="true"></span>
              <span>{{ item.label }}</span>
            </a>
          }

          <div class="nav-divider"></div>
          <a routerLink="/donantes/registro" (click)="closeMobileMenu()">
            <span class="nav-dot" aria-hidden="true"></span>
            <span>Portal de donantes</span>
          </a>
        </nav>

        @if (auth.currentUser(); as user) {
          <div class="sidebar-user">
            <div class="user-name">{{ user.nombreCompleto }}</div>
            <div class="user-role">{{ formatRolLabel(user.rol) }}</div>
            <div class="user-hospital">{{ formatHospitalLabel(user.hospitalId) }}</div>
            <button type="button" class="logout-btn" (click)="auth.logout()">
              Cerrar sesión
            </button>
          </div>
        }
      </aside>

      <div class="workspace-content">
        <main class="workspace-main">
          <router-outlet />
        </main>
      </div>
    </div>
  `,
  styles: [`
    .workspace-shell {
      min-height: 100vh;
      display: flex;
      background-color: var(--bg-canvas);
    }
    .sidebar {
      width: 262px;
      margin: 1rem 0 1rem 1rem;
      border-radius: var(--radius-card);
      background-color: var(--brand-ink);
      color: var(--text-on-ink);
      display: flex;
      flex-direction: column;
      justify-content: space-between;
      padding: 1.5rem 1.15rem;
      flex-shrink: 0;
      box-shadow: var(--shadow-soft);
    }
    .sidebar-top {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 0.35rem;
    }
    .sidebar-brand {
      display: inline-block;
      font-size: 1.45rem;
      font-weight: 900;
      letter-spacing: -0.02em;
      line-height: 1.1;
      color: var(--text-on-ink);
    }
    .mobile-toggle {
      display: none;
      background: rgba(250, 248, 242, 0.12);
      color: var(--text-on-ink);
      border: 1px solid rgba(250, 248, 242, 0.25);
      padding: 0.4rem 0.9rem;
      font-size: 0.82rem;
      font-weight: 700;
      border-radius: var(--radius-pill);
      cursor: pointer;
    }
    .sidebar-nav {
      display: flex;
      flex-direction: column;
      gap: 0.35rem;
      margin: 1.75rem 0 auto;
    }
    .sidebar-nav a {
      display: flex;
      align-items: center;
      gap: 0.65rem;
      color: rgba(250, 248, 242, 0.82);
      padding: 0.65rem 1rem;
      border-radius: var(--radius-pill);
      font-size: 0.9rem;
      font-weight: 700;
    }
    .nav-dot {
      width: 7px;
      height: 7px;
      border-radius: 50%;
      background-color: #3AA1B3;
      flex-shrink: 0;
    }
    .sidebar-nav a:hover {
      background-color: rgba(250, 248, 242, 0.1);
      color: #FFFFFF;
      opacity: 1;
    }
    .sidebar-nav a.active {
      background-color: var(--brand-yellow);
      color: var(--brand-ink);
      font-weight: 800;
    }
    .sidebar-nav a.active .nav-dot {
      background-color: var(--brand-ink);
    }
    .nav-divider {
      height: 1px;
      background-color: rgba(250, 248, 242, 0.14);
      margin: 0.65rem 0.5rem;
    }
    .sidebar-user {
      padding: 1.1rem;
      border-radius: var(--radius-subcard);
      background-color: rgba(250, 248, 242, 0.08);
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
    }
    .user-name {
      font-size: 0.92rem;
      font-weight: 800;
      color: #FFFFFF;
    }
    .user-role {
      font-size: 0.8rem;
      font-weight: 700;
      color: var(--brand-yellow);
    }
    .user-hospital {
      font-size: 0.78rem;
      font-weight: 600;
      color: rgba(250, 248, 242, 0.78);
    }
    .logout-btn {
      margin-top: 0.65rem;
      padding: 0.5rem 0.9rem;
      background: rgba(250, 248, 242, 0.12);
      color: #FFFFFF;
      border: none;
      border-radius: var(--radius-pill);
      font-size: 0.82rem;
      font-weight: 700;
      cursor: pointer;
    }
    .logout-btn:hover {
      background-color: rgba(250, 248, 242, 0.22);
    }
    .workspace-content { flex: 1; min-width: 0; }
    .workspace-main {
      max-width: 1240px;
      margin: 0 auto;
      padding: 2.25rem 2.25rem 3.5rem;
    }
    @media (max-width: 900px) {
      .workspace-shell { flex-direction: column; }
      .sidebar {
        width: auto;
        margin: 0.75rem;
        border-radius: var(--radius-subcard);
        padding: 1rem 1.15rem;
      }
      .mobile-toggle { display: inline-block; }
      .sidebar-nav, .sidebar-user { display: none; }
      .sidebar.open .sidebar-nav, .sidebar.open .sidebar-user { display: flex; }
      .workspace-main { padding: 1.25rem 1rem 2.5rem; }
    }
  `],
})
export class WorkspaceLayoutComponent {
  protected readonly auth = inject(AuthService);
  protected readonly mobileMenuOpen = signal(false);
  protected readonly formatRolLabel = formatRol;
  protected readonly formatHospitalLabel = formatHospital;

  protected readonly visibleNavItems = computed(() => {
    const rol = this.auth.currentRole();
    if (!rol) return [];
    return WORKSPACE_NAV_ITEMS.filter((item) => item.roles.includes(rol));
  });

  toggleMobileMenu(): void {
    this.mobileMenuOpen.update((v) => !v);
  }

  closeMobileMenu(): void {
    this.mobileMenuOpen.set(false);
  }
}
