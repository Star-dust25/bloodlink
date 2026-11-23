import { ChangeDetectionStrategy, Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { DashboardMetricas, StockPorGrupoKPI } from '../../core/models/reporte.model';
import { AuthService } from '../../core/services/auth.service';
import { ReportesService } from '../../core/services/reportes.service';
import { ExtractedError, extractApiError } from '../../core/utils/error.util';
import { formatEstadoSolicitud } from '../../core/utils/label.util';
import { InlineAlertComponent } from '../../shared/ui/inline-alert/inline-alert.component';
import { PageHeaderComponent } from '../../shared/ui/page-header/page-header.component';
import { StatusIndicatorComponent } from '../../shared/ui/status-indicator/status-indicator.component';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [RouterLink, PageHeaderComponent, InlineAlertComponent, StatusIndicatorComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <app-page-header
      eyebrow="RESUMEN DE LA RED"
      title="Resumen general de la red"
      subtitle="Estado actual de las unidades disponibles, emergencias atendidas y donantes voluntarios en la región."
    >
      <button type="button" class="btn btn-ghost" (click)="cargarDashboard()">
        Actualizar datos
      </button>
    </app-page-header>

    @if (apiError(); as err) {
      <div style="margin-bottom: 1.5rem;">
        <app-inline-alert variant="error" [title]="err.mensaje" [message]="err.detalle" />
      </div>
    }

    @if (metricas(); as m) {
      <div class="kpi-grid tabular">
        <div class="kpi-card kpi-turquoise">
          <span class="eyebrow">UNIDADES DISPONIBLES</span>
          <div class="kpi-value">{{ m.totalUnidadesDisponibles }}</div>
          <span class="kpi-sub">En {{ m.totalBancosRed }} hospitales conectados</span>
        </div>

        <div class="kpi-card kpi-yellow">
          <span class="eyebrow">POR VENCER (7 DÍAS)</span>
          <div class="kpi-value">{{ m.unidadesProximasACaducar7d }}</div>
          <span class="kpi-sub">
            <app-status-indicator
              [tone]="m.unidadesAlertaRoja3d > 0 ? 'critical' : m.unidadesProximasACaducar7d > 0 ? 'warning' : 'ok'"
              [label]="m.unidadesAlertaRoja3d + ' en vencimiento crítico'"
            />
          </span>
        </div>

        <div class="kpi-card">
          <span class="eyebrow">EMERGENCIAS ATENDIDAS</span>
          <div class="kpi-value">{{ m.emergenciasAtendidas }} / {{ m.emergenciasActivas }}</div>
          <span class="kpi-sub">Atendidas / activas ({{ m.totalEmergencias }} en total)</span>
        </div>

        <div class="kpi-card">
          <span class="eyebrow">TIEMPO DE RESPUESTA</span>
          <div class="kpi-value">{{ m.tiempoPromedioMatchingMs }} ms</div>
          <span class="kpi-sub">Promedio de ubicación de stock</span>
        </div>

        <div class="kpi-card">
          <span class="eyebrow">DONANTES VOLUNTARIOS</span>
          <div class="kpi-value">{{ m.donantesHabilitados }}</div>
          <span class="kpi-sub">Listos para recibir avisos</span>
        </div>
      </div>

      <div class="dashboard-columns">
        <!-- GRÁFICA DE BARRAS REDONDEADAS EN CSS PURO -->
        <section class="surface-card">
          <div class="sec-head">
            <span class="eyebrow">INVENTARIO REGIONAL</span>
            <h2 class="sec-title">Unidades por grupo sanguíneo</h2>
          </div>

          <div class="bars-list tabular">
            @for (item of m.stockPorTipoSangre; track item.grupoSanguineo) {
              <div class="bar-row">
                <div class="bar-label">
                  <strong>{{ item.grupoSanguineo }}</strong>
                  <span class="bar-counts">
                    {{ item.unidadesDisponibles }} disponibles
                    @if (item.unidadesReservadas > 0) {
                      · {{ item.unidadesReservadas }} reservadas
                    }
                  </span>
                </div>
                <div class="bar-track">
                  <div
                    class="bar-fill"
                    [class.bar-critical]="item.unidadesDisponibles === 0"
                    [style.width.%]="barPercentage(item)"
                  ></div>
                </div>
              </div>
            }
          </div>
        </section>

        <div class="side-stack">
          <section class="surface-card">
            <div class="sec-head">
              <span class="eyebrow">SOLICITUDES</span>
              <h2 class="sec-title">Emergencias por estado</h2>
            </div>

            <div class="state-list tabular">
              @for (est of m.emergenciasPorEstado; track est.estado) {
                <div class="state-row">
                  <span>{{ formatEstado(est.estado) }}</span>
                  <strong>{{ est.cantidad }}</strong>
                </div>
              }
            </div>
          </section>

          <section class="surface-card">
            <span class="eyebrow">ACCESOS RÁPIDOS</span>
            <h2 class="sec-title" style="margin-bottom: 0.85rem;">¿Qué quieres revisar?</h2>
            <div class="quick-links">
              <a routerLink="/emergencias" class="quick-link">
                <span>Solicitudes de emergencia</span>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <path d="M9 18L15 12L9 6" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </a>
              <a routerLink="/inventario" class="quick-link">
                <span>Stock en red e inventario</span>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <path d="M9 18L15 12L9 6" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </a>
              @if (puedeAuditar()) {
                <a routerLink="/trazabilidad" class="quick-link">
                  <span>Historial y certificados PDF / CSV</span>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                    <path d="M9 18L15 12L9 6" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                </a>
                <a routerLink="/donantes" class="quick-link">
                  <span>Donantes voluntarios</span>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                    <path d="M9 18L15 12L9 6" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                </a>
              }
              <a routerLink="/notificaciones" class="quick-link">
                <span>Alertas y avisos ({{ m.ultimasNotificaciones.length }})</span>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <path d="M9 18L15 12L9 6" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </a>
            </div>
          </section>
        </div>
      </div>
    }
  `,
  styles: [`
    .kpi-turquoise { border-top: 4px solid var(--brand-turquoise); }
    .kpi-yellow { border-top: 4px solid var(--brand-yellow); }
    .dashboard-columns { display: grid; grid-template-columns: 1.35fr 1fr; gap: 1.5rem; }
    .side-stack { display: flex; flex-direction: column; gap: 1.5rem; }
    .sec-head { margin-bottom: 1.25rem; }
    .sec-title { font-size: 1.2rem; font-weight: 800; margin-top: 0.25rem; }
    .bars-list { display: flex; flex-direction: column; gap: 1.05rem; }
    .bar-row { display: flex; flex-direction: column; gap: 0.4rem; }
    .bar-label { display: flex; justify-content: space-between; font-size: 0.9rem; }
    .bar-counts { color: var(--text-muted); font-size: 0.82rem; font-weight: 600; }
    .bar-track {
      width: 100%; height: 14px; background-color: var(--bg-surface-muted);
      border-radius: var(--radius-pill); overflow: hidden;
    }
    .bar-fill {
      height: 100%; border-radius: var(--radius-pill);
      background: linear-gradient(90deg, var(--brand-turquoise) 0%, var(--brand-sky) 100%);
      transition: width 300ms ease;
    }
    .bar-fill.bar-critical {
      width: 5% !important;
      background: var(--state-danger);
    }
    .state-list { display: flex; flex-direction: column; }
    .state-row {
      display: flex; justify-content: space-between; padding: 0.75rem 0;
      border-bottom: 1px solid var(--border-subtle); font-size: 0.92rem; font-weight: 600;
    }
    .state-row:last-child { border-bottom: none; }
    .quick-links { display: flex; flex-direction: column; gap: 0.6rem; }
    .quick-link {
      display: flex; justify-content: space-between; align-items: center;
      padding: 0.85rem 1.15rem; background-color: var(--bg-surface-muted);
      border: 1.5px solid var(--border-subtle); border-radius: var(--radius-pill);
      font-size: 0.88rem; font-weight: 700;
    }
    .quick-link:hover {
      border-color: var(--brand-turquoise);
      background-color: var(--brand-turquoise-soft);
      opacity: 1;
    }
    @media (max-width: 860px) {
      .dashboard-columns { grid-template-columns: 1fr; }
    }
  `],
})
export class DashboardComponent implements OnInit {
  private readonly reportesService = inject(ReportesService);
  private readonly authService = inject(AuthService);

  protected readonly metricas = signal<DashboardMetricas | null>(null);
  protected readonly apiError = signal<ExtractedError | null>(null);
  protected readonly formatEstado = formatEstadoSolicitud;
  protected readonly puedeAuditar = computed(() =>
    this.authService.hasRole(['COORDINADOR', 'PERSONAL_BANCO'])
  );

  ngOnInit(): void {
    this.cargarDashboard();
  }

  cargarDashboard(): void {
    this.apiError.set(null);
    this.reportesService.obtenerDashboard().subscribe({
      next: (data) => this.metricas.set(data),
      error: (err) => this.apiError.set(extractApiError(err)),
    });
  }

  barPercentage(item: StockPorGrupoKPI): number {
    const maxRef = 20;
    if (item.unidadesDisponibles <= 0) return 5;
    return Math.min(100, Math.round((item.unidadesDisponibles / maxRef) * 100));
  }
}
