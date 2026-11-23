import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../core/services/auth.service';
import { extractApiError } from '../../core/utils/error.util';

interface DemoAccount {
  rolHumano: string;
  email: string;
  password: string;
}

const DEMO_ACCOUNTS: readonly DemoAccount[] = [
  {
    rolHumano: 'Médico',
    email: 'medico@bloodlink.gob.pe',
    password: 'Medico123!',
  },
  {
    rolHumano: 'Personal de Banco',
    email: 'banco@bloodlink.gob.pe',
    password: 'Banco123!',
  },
  {
    rolHumano: 'Coordinador',
    email: 'coordinador@bloodlink.gob.pe',
    password: 'Coord123!',
  },
] as const;

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [FormsModule, RouterLink],
  changeDetection: ChangeDetectionStrategy.OnPush,
  host: {
    '(document:click)': 'cerrarTooltip()',
    '(document:keydown.escape)': 'cerrarTooltip()',
  },
  template: `
    <div class="login-page">
      <header class="login-topbar">
        <a routerLink="/" class="login-brand">
          <span class="brand-dot" aria-hidden="true"></span>
          <span>BloodLink</span>
        </a>
        <a routerLink="/" class="back-link">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M15 18L9 12L15 6" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
          <span>Volver</span>
        </a>
      </header>

      <main class="login-card">
        <div class="login-header">
          <h1 class="login-title">Iniciar sesión</h1>
          <p class="login-sub">Acceso para personal de salud.</p>
        </div>

        <form class="form-stack" novalidate (ngSubmit)="onSubmit()">
          <div class="field">
            <div class="label-with-tip">
              <label class="field-label" for="emailInput">Correo electrónico</label>

              <div class="tip-anchor" [class.open]="tooltipCuentasAbierto()" (click)="$event.stopPropagation()">
                <button
                  type="button"
                  class="tip-btn"
                  aria-label="Ver cuentas de prueba"
                  [attr.aria-expanded]="tooltipCuentasAbierto()"
                  (click)="toggleTooltipCuentas($event)"
                >
                  ?
                </button>

                <div class="tip-popover tip-popover-wide" role="tooltip">
                  <div class="tip-head">Cuentas de prueba</div>
                  @for (acc of demoAccounts; track acc.email) {
                    <button
                      type="button"
                      class="tip-option"
                      (click)="seleccionarCuentaPrueba(acc)"
                    >
                      <strong>{{ acc.rolHumano }}</strong> · {{ acc.email }}
                    </button>
                  }
                </div>
              </div>
            </div>

            <input
              id="emailInput"
              name="email"
              type="email"
              autocomplete="email"
              class="input login-input"
              [class.input-error]="!!errorEmail()"
              placeholder="nombre@bloodlink.gob.pe"
              required
              [ngModel]="email"
              (ngModelChange)="onEmailChange($event)"
            />
            @if (errorEmail(); as errMail) {
              <p class="field-error" role="alert">{{ errMail }}</p>
            }
          </div>

          <div class="field">
            <label class="field-label" for="passwordInput">Contraseña</label>
            <div class="password-wrap">
              <input
                id="passwordInput"
                name="password"
                [type]="mostrarPassword() ? 'text' : 'password'"
                autocomplete="current-password"
                class="input login-input password-input"
                [class.input-error]="!!errorPassword()"
                placeholder="••••••••••"
                required
                [ngModel]="password"
                (ngModelChange)="onPasswordChange($event)"
              />
              <button
                type="button"
                class="pwd-toggle"
                [attr.aria-label]="mostrarPassword() ? 'Ocultar contraseña' : 'Mostrar contraseña'"
                [attr.aria-pressed]="mostrarPassword()"
                (click)="toggleMostrarPassword()"
              >
                @if (mostrarPassword()) {
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                    <path d="M3 3L21 21" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
                    <path d="M10.58 10.58A2 2 0 0 0 13.42 13.42" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
                    <path d="M9.88 5.09A10.94 10.94 0 0 1 12 4.9C17 4.9 20.73 8.11 22 12C21.55 13.38 20.82 14.62 19.87 15.65M6.61 6.61C4.62 7.96 3.06 9.81 2 12C3.27 15.89 7 19.1 12 19.1C13.91 19.1 15.68 18.63 17.2 17.8" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                } @else {
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                    <path d="M2 12C3.27 8.11 7 4.9 12 4.9C17 4.9 20.73 8.11 22 12C20.73 15.89 17 19.1 12 19.1C7 19.1 3.27 15.89 2 12Z" stroke="currentColor" stroke-width="2" stroke-linejoin="round" />
                    <circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="2" />
                  </svg>
                }
              </button>
            </div>
            @if (errorPassword(); as errPwd) {
              <p class="field-error" role="alert">{{ errPwd }}</p>
            }
          </div>

          @if (serverError(); as srvErr) {
            <p class="field-error" role="alert">{{ srvErr }}</p>
          }

          <button
            type="submit"
            class="btn btn-primary btn-full login-submit"
            [disabled]="loading()"
          >
            {{ loading() ? 'Ingresando...' : 'Ingresar' }}
          </button>
        </form>
      </main>
    </div>
  `,
  styles: [`
    .login-page {
      min-height: 100dvh;
      background-color: var(--bg-canvas);
      padding: 2rem 1.25rem;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 1rem;
    }
    .login-topbar {
      width: 100%;
      max-width: 480px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 0.25rem;
    }
    .login-brand {
      display: inline-flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 1.2rem;
      font-weight: 900;
    }
    .brand-dot {
      width: 12px;
      height: 12px;
      border-radius: 50%;
      background: linear-gradient(135deg, var(--brand-turquoise) 50%, var(--brand-yellow) 50%);
    }
    .back-link {
      display: inline-flex;
      align-items: center;
      gap: 0.35rem;
      font-size: 0.94rem;
      font-weight: 800;
      color: var(--brand-turquoise-dark);
    }
    .login-card {
      width: 100%;
      max-width: 480px;
      background-color: var(--bg-surface);
      border: 1px solid var(--border-subtle);
      border-radius: 32px;
      padding: 2.25rem;
      box-shadow: var(--shadow-soft);
      display: flex;
      flex-direction: column;
      gap: 1.4rem;
    }
    .login-header {
      display: flex;
      flex-direction: column;
      gap: 0.35rem;
    }
    .login-title {
      font-size: 1.85rem;
      font-weight: 900;
      color: var(--brand-ink);
    }
    .login-sub {
      font-size: 0.98rem;
      color: var(--text-secondary);
    }
    .label-with-tip {
      display: flex;
      align-items: center;
      justify-content: flex-start;
      gap: 0.5rem;
    }
    .tip-popover-wide {
      width: max-content;
      max-width: min(350px, calc(100vw - 3.5rem));
    }
    .login-input {
      min-height: 54px;
    }
    .password-wrap {
      position: relative;
      display: flex;
      align-items: center;
    }
    .password-input {
      padding-right: 3rem;
    }
    .pwd-toggle {
      position: absolute;
      right: 0.75rem;
      width: 36px;
      height: 36px;
      border: none;
      border-radius: 50%;
      background: transparent;
      color: var(--text-secondary);
      display: inline-flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;
      transition: background-color 180ms ease, color 180ms ease;
    }
    .pwd-toggle:hover {
      background-color: var(--bg-surface-muted);
      color: var(--brand-ink);
    }
    .pwd-toggle:focus-visible {
      outline: 2px solid var(--brand-turquoise);
      outline-offset: 2px;
    }
    .btn-full {
      width: 100%;
    }
    .login-submit {
      min-height: 54px;
      font-weight: 800;
    }
    @media (max-width: 520px) {
      .login-card {
        padding: 1.5rem;
      }
    }
  `],
})
export class LoginComponent {
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);

  protected readonly demoAccounts = DEMO_ACCOUNTS;
  protected readonly loading = signal(false);
  protected readonly tooltipCuentasAbierto = signal(false);
  protected readonly mostrarPassword = signal(false);
  protected readonly errorEmail = signal('');
  protected readonly errorPassword = signal('');
  protected readonly serverError = signal('');

  protected email = '';
  protected password = '';

  toggleTooltipCuentas(event: Event): void {
    event.stopPropagation();
    this.tooltipCuentasAbierto.update((v) => !v);
  }

  cerrarTooltip(): void {
    if (this.tooltipCuentasAbierto()) {
      this.tooltipCuentasAbierto.set(false);
    }
  }

  seleccionarCuentaPrueba(acc: DemoAccount): void {
    this.email = acc.email;
    this.password = acc.password;
    this.errorEmail.set('');
    this.errorPassword.set('');
    this.serverError.set('');
    this.tooltipCuentasAbierto.set(false);
  }

  toggleMostrarPassword(): void {
    this.mostrarPassword.update((v) => !v);
  }

  onEmailChange(value: string): void {
    this.email = value ?? '';
    if (this.errorEmail()) this.errorEmail.set('');
    if (this.serverError()) this.serverError.set('');
  }

  onPasswordChange(value: string): void {
    this.password = value ?? '';
    if (this.errorPassword()) this.errorPassword.set('');
    if (this.serverError()) this.serverError.set('');
  }

  onSubmit(): void {
    this.cerrarTooltip();

    const cleanEmail = this.email.trim();
    const emailValid = EMAIL_REGEX.test(cleanEmail);
    const passwordValid = this.password.length > 0;

    this.errorEmail.set(emailValid ? '' : 'Ingresa un correo electrónico válido.');
    this.errorPassword.set(passwordValid ? '' : 'Ingresa tu contraseña.');

    if (!emailValid || !passwordValid) {
      return;
    }

    this.loading.set(true);
    this.serverError.set('');

    this.authService.login({ email: cleanEmail, password: this.password }).subscribe({
      next: (resp) => {
        this.loading.set(false);
        const target = this.authService.defaultRouteForRole(resp.usuario.rol);
        void this.router.navigate([target]);
      },
      error: (err) => {
        this.loading.set(false);
        const parsed = extractApiError(err);
        if (parsed.status === 401) {
          this.serverError.set('Correo o contraseña incorrectos.');
          return;
        }
        if (parsed.status === 423) {
          this.serverError.set(
            'Cuenta bloqueada temporalmente por intentos fallidos. Intenta nuevamente en 15 minutos.'
          );
          return;
        }
        const raw = (parsed.mensaje || 'No pudimos iniciar sesión. Verifica tus datos.').trim();
        const cap = raw.charAt(0).toUpperCase() + raw.slice(1);
        this.serverError.set(/[.!?]$/.test(cap) ? cap : `${cap}.`);
      },
    });
  }
}
