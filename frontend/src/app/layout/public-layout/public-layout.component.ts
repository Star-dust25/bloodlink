import { ChangeDetectionStrategy, Component, DestroyRef, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { NavigationEnd, Router, RouterLink, RouterOutlet } from '@angular/router';
import { filter } from 'rxjs';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-public-layout',
  standalone: true,
  imports: [RouterOutlet, RouterLink],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="public-shell">
      <header class="public-header no-print">
        <div class="header-pill">
          <a routerLink="/" class="brand" (click)="irAlInicio($event)">
            <span class="brand-dot" aria-hidden="true"></span>
            <span class="brand-name">BloodLink</span>
          </a>

          <nav class="public-nav" aria-label="Secciones principales">
            <a href="#como-funciona" (click)="navegarASeccion($event, 'como-funciona')">Cómo funciona</a>
            <a href="#compatibilidad" (click)="navegarASeccion($event, 'compatibilidad')">Compatibilidad</a>
            <a href="#hospitales" (click)="navegarASeccion($event, 'hospitales')">Hospitales</a>
            <a href="#mitos" (click)="navegarASeccion($event, 'mitos')">Mitos y dudas</a>
            <a href="#requisitos" (click)="navegarASeccion($event, 'requisitos')">Requisitos</a>
          </nav>

          @if (auth.isAuthenticated()) {
            <a
              [routerLink]="auth.defaultRouteForRole(auth.currentRole())"
              class="btn btn-secondary btn-sm"
            >
              Ir a mi panel
            </a>
          } @else {
            <a routerLink="/login" class="btn btn-secondary btn-sm">
              Acceso hospitales
            </a>
          }
        </div>
      </header>

      <main class="public-main">
        <router-outlet />
      </main>

      <footer class="public-footer no-print">
        <div class="footer-inner">
          <span>BloodLink · Red de Bancos de Sangre — Piura, Perú</span>
          <div class="footer-links">
            <span>Tus datos están protegidos</span>
            <span>·</span>
            <a routerLink="/interoperabilidad" class="footer-subtle-link">Integraciones</a>
          </div>
        </div>
      </footer>
    </div>
  `,
  styles: [`
    .public-shell {
      min-height: 100vh;
      display: flex;
      flex-direction: column;
      background-color: var(--bg-canvas);
    }
    .public-header {
      position: sticky;
      top: 14px;
      z-index: 100;
      padding: 0.75rem 1.5rem 0.25rem;
    }
    .header-pill {
      max-width: 1240px;
      margin: 0 auto;
      padding: 0.7rem 1.35rem;
      background-color: rgba(255, 255, 255, 0.9);
      backdrop-filter: blur(12px);
      -webkit-backdrop-filter: blur(12px);
      border: 1px solid var(--border-subtle);
      border-radius: var(--radius-pill);
      box-shadow: 0 12px 32px rgba(34, 30, 30, 0.08);
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      justify-content: space-between;
      gap: 1rem;
    }
    .brand {
      display: inline-flex;
      align-items: center;
      gap: 0.55rem;
    }
    .brand-dot {
      width: 14px;
      height: 14px;
      border-radius: 50%;
      background: linear-gradient(135deg, var(--brand-turquoise) 50%, var(--brand-yellow) 50%);
      flex-shrink: 0;
    }
    .brand-name {
      font-size: 1.28rem;
      font-weight: 900;
      letter-spacing: -0.025em;
      color: var(--brand-ink);
    }
    .public-nav {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      gap: 1.25rem;
      font-size: 0.92rem;
      font-weight: 700;
    }
    .public-nav a {
      color: var(--text-secondary);
      padding: 0.3rem 0.45rem;
    }
    .public-nav a:hover {
      color: var(--brand-turquoise-dark);
      opacity: 1;
    }
    .public-main {
      flex: 1;
      width: 100%;
      max-width: 1240px;
      margin: 0 auto;
      padding: 0.75rem 1.5rem 4rem;
    }
    .public-footer {
      border-top: 1px solid var(--border-subtle);
      padding: 1.35rem 1.5rem;
      font-size: 0.88rem;
      font-weight: 600;
      color: var(--text-muted);
    }
    .footer-inner {
      max-width: 1240px;
      margin: 0 auto;
      display: flex;
      flex-wrap: wrap;
      justify-content: space-between;
      gap: 0.75rem;
    }
    .footer-links {
      display: flex;
      align-items: center;
      gap: 0.6rem;
    }
    .footer-subtle-link {
      color: var(--text-muted);
      text-decoration: underline;
      text-underline-offset: 3px;
    }
    @media (max-width: 880px) {
      .public-nav {
        display: none;
      }
    }
  `],
})
export class PublicLayoutComponent {
  protected readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);
  private previousUrl = '';

  constructor() {
    this.router.events
      .pipe(
        filter((ev): ev is NavigationEnd => ev instanceof NavigationEnd),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe((ev) => {
        const cleanUrl = ev.urlAfterRedirects.split('#')[0];
        if (this.previousUrl && cleanUrl !== this.previousUrl) {
          const state = typeof window !== 'undefined' ? (window.history.state as { scrollTo?: string } | null) : null;
          if (!state?.scrollTo && typeof window !== 'undefined') {
            window.scrollTo({ top: 0, behavior: 'auto' });
          }
        }
        this.previousUrl = cleanUrl;
      });
  }

  irAlInicio(event: Event): void {
    const basePath = this.router.url.split('#')[0];
    if (basePath === '/' || basePath === '') {
      event.preventDefault();
      if (typeof window !== 'undefined') {
        window.scrollTo({ top: 0, behavior: 'smooth' });
        window.history.replaceState({}, '', '/');
      }
    }
  }

  navegarASeccion(event: Event, seccionId: string): void {
    event.preventDefault();
    const basePath = this.router.url.split('#')[0];
    if (basePath === '/' || basePath === '') {
      if (typeof document !== 'undefined') {
        document.getElementById(seccionId)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }
      if (typeof window !== 'undefined') {
        window.history.replaceState({}, '', '/');
      }
      return;
    }
    void this.router.navigate(['/'], { state: { scrollTo: seccionId } });
  }
}
