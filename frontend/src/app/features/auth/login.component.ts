import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../../core/services/auth.service';
import { ExtractedError, extractApiError } from '../../core/utils/error.util';
import { formatRol } from '../../core/utils/label.util';
import { InlineAlertComponent } from '../../shared/ui/inline-alert/inline-alert.component';

interface DemoAccount {
  label: string;
  rol: string;
  email: string;
  password: string;
}

const DEMO_ACCOUNTS: readonly DemoAccount[] = [
  {
    label: 'Dr. Carlos Mendoza',
    rol: 'MEDICO',
    email: 'medico@bloodlink.gob.pe',
    password: 'Medico123!',
  },
  {
    label: 'Lic. María Torres',
    rol: 'PERSONAL_BANCO',
    email: 'banco@bloodlink.gob.pe',
    password: 'Banco123!',
  },
  {
    label: 'Mg. Roberto Paredes',
    rol: 'COORDINADOR',
    email: 'coordinador@bloodlink.gob.pe',
    password: 'Coord123!',
  },
];

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [FormsModule, InlineAlertComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="login-wrapper">
      <section class="login-card surface-card">
        <div class="login-header">
          <span class="eyebrow">PORTAL DE SALUD</span>
          <h1 class="login-title">Bienvenido de nuevo</h1>
          <p class="login-subtitle">
            Ingresa con tu correo institucional para acceder a tu panel de trabajo.
          </p>
        </div>

        @if (apiError(); as err) {
          <app-inline-alert
            variant="error"
            [title]="err.status === 423 ? 'Cuenta bloqueada temporalmente (15 minutos)' : err.mensaje"
            [message]="err.detalle || (err.status === 423 ? 'Se superó el límite de 3 intentos fallidos consecutivos.' : '')"
          />
        }

        <form class="form-stack" (ngSubmit)="onSubmit()">
          <div class="field">
            <label class="field-label" for="email">Correo institucional</label>
            <input
              id="email"
              name="email"
              type="email"
              class="input"
              required
              autocomplete="username"
              placeholder="usuario@bloodlink.gob.pe"
              [(ngModel)]="email"
            />
          </div>

          <div class="field">
            <label class="field-label" for="password">Contraseña</label>
            <input
              id="password"
              name="password"
              type="password"
              class="input"
              required
              autocomplete="current-password"
              placeholder="••••••••••"
              [(ngModel)]="password"
            />
          </div>

          <button
            type="submit"
            class="btn btn-primary btn-full"
            [disabled]="loading() || !email.trim() || !password"
          >
            {{ loading() ? 'Verificando tus datos...' : 'Ingresar a mi panel' }}
          </button>
        </form>

        <details class="demo-details">
          <summary class="demo-summary">Cuentas de prueba rápida</summary>
          <div class="demo-list">
            @for (acc of demoAccounts; track acc.email) {
              <button
                type="button"
                class="demo-item"
                (click)="fillDemoAccount(acc)"
              >
                <div>
                  <div class="demo-role">{{ formatRolLabel(acc.rol) }}</div>
                  <div class="demo-name">{{ acc.label }}</div>
                </div>
                <span class="demo-action">Usar</span>
              </button>
            }
          </div>
        </details>
      </section>
    </div>
  `,
  styles: [`
    .login-wrapper {
      display: flex;
      justify-content: center;
      padding: 1.5rem 0;
    }
    .login-card {
      width: 100%;
      max-width: 470px;
      padding: 2.5rem;
      border-top: 5px solid var(--brand-turquoise);
      display: flex;
      flex-direction: column;
      gap: 1.5rem;
    }
    .login-header {
      display: flex;
      flex-direction: column;
      gap: 0.4rem;
    }
    .login-title { font-size: 1.85rem; font-weight: 900; }
    .login-subtitle { font-size: 0.94rem; color: var(--text-secondary); }
    .btn-full { width: 100%; margin-top: 0.25rem; }
    .demo-details {
      padding-top: 1.25rem;
      border-top: 1px solid var(--border-subtle);
    }
    .demo-summary {
      font-size: 0.84rem;
      font-weight: 800;
      color: var(--brand-turquoise-dark);
      cursor: pointer;
      user-select: none;
    }
    .demo-list {
      margin-top: 0.85rem;
      display: flex;
      flex-direction: column;
      gap: 0.55rem;
    }
    .demo-item {
      width: 100%;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0.75rem 1rem;
      background-color: var(--bg-surface-muted);
      border: 1.5px solid var(--border-subtle);
      border-radius: var(--radius-input);
      cursor: pointer;
      text-align: left;
      transition: all var(--transition-fast);
    }
    .demo-item:hover {
      border-color: var(--brand-turquoise);
      background-color: var(--brand-turquoise-soft);
    }
    .demo-role { font-size: 0.78rem; font-weight: 800; color: var(--brand-turquoise-dark); }
    .demo-name { font-size: 0.9rem; font-weight: 700; color: var(--text-primary); }
    .demo-action { font-size: 0.84rem; font-weight: 800; color: var(--brand-ink); }
  `],
})
export class LoginComponent {
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);

  protected readonly demoAccounts = DEMO_ACCOUNTS;
  protected readonly formatRolLabel = formatRol;
  protected email = '';
  protected password = '';
  protected readonly loading = signal(false);
  protected readonly apiError = signal<ExtractedError | null>(null);

  fillDemoAccount(acc: DemoAccount): void {
    this.email = acc.email;
    this.password = acc.password;
    this.apiError.set(null);
  }

  onSubmit(): void {
    if (!this.email.trim() || !this.password) return;
    this.loading.set(true);
    this.apiError.set(null);

    this.authService
      .login({ email: this.email.trim(), password: this.password })
      .subscribe({
        next: (resp) => {
          this.loading.set(false);
          const target = this.authService.defaultRouteForRole(resp.usuario.rol);
          void this.router.navigate([target]);
        },
        error: (err) => {
          this.loading.set(false);
          this.apiError.set(extractApiError(err));
        },
      });
  }
}
