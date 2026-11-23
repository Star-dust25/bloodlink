import { ChangeDetectionStrategy, Component, OnInit, computed, inject, input, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { EmergenciaMatchingResponseDTO } from '../../core/models/emergencia.model';
import { AuthService } from '../../core/services/auth.service';
import { EmergenciasService } from '../../core/services/emergencias.service';
import { ExtractedError, extractApiError } from '../../core/utils/error.util';
import {
  formatEstadoSolicitud,
  formatOrigenAsignacion,
  formatUrgencia,
} from '../../core/utils/label.util';
import { InlineAlertComponent } from '../../shared/ui/inline-alert/inline-alert.component';
import { PageHeaderComponent } from '../../shared/ui/page-header/page-header.component';
import { StatusIndicatorComponent } from '../../shared/ui/status-indicator/status-indicator.component';

@Component({
  selector: 'app-emergencia-detalle',
  standalone: true,
  imports: [
    FormsModule,
    RouterLink,
    PageHeaderComponent,
    InlineAlertComponent,
    StatusIndicatorComponent,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <app-page-header
      eyebrow="DETALLE DE EMERGENCIA"
      [title]="solicitud()?.solicitudId || solicitudId()"
      subtitle="Información de las unidades reservadas, banco asignado y estado de la atención."
    >
      <a routerLink="/emergencias" class="btn btn-ghost">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
          <path d="M15 18L9 12L15 6" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
        <span>Volver a Emergencias</span>
      </a>
    </app-page-header>

    @if (apiError(); as err) {
      <div style="margin-bottom: 1.5rem;">
        <app-inline-alert variant="error" [title]="err.mensaje" [message]="err.detalle" />
      </div>
    }

    @if (accionExitosa(); as msg) {
      <div style="margin-bottom: 1.5rem;">
        <app-inline-alert variant="success" title="Actualización guardada" [message]="msg" />
      </div>
    }

    @if (solicitud(); as sol) {
      @if (sol.fallbackJurisdiccionUsado) {
        <div style="margin-bottom: 1.5rem;">
          <app-inline-alert
            variant="warning"
            title="Búsqueda en modo respaldo por zona"
            message="Como el servicio de mapas no estuvo disponible en este momento, seleccionamos el banco priorizando la misma zona o distrito."
          />
        </div>
      }

      <div class="kpi-grid tabular">
        <div class="kpi-card">
          <span class="eyebrow">ESTADO ACTUAL</span>
          <div style="margin-top: 0.4rem;">
            <app-status-indicator
              [tone]="sol.estadoSolicitud === 'Atendida' ? 'ok' : sol.estadoSolicitud === 'Reservada' ? 'ink' : sol.estadoSolicitud === 'En Espera de Donantes' ? 'critical' : 'neutral'"
              [label]="formatEstado(sol.estadoSolicitud)"
            />
          </div>
          <span class="kpi-sub">{{ sol.fechaHora }}</span>
        </div>

        <div class="kpi-card">
          <span class="eyebrow">UNIDADES REQUERIDAS</span>
          <div class="kpi-value">{{ sol.tipoSangreReq }}{{ sol.factorRhReq }}</div>
          <span class="kpi-sub">{{ sol.cantidadReq }} unidades · {{ formatUrg(sol.urgencia) }}</span>
        </div>

        <div class="kpi-card">
          <span class="eyebrow">ORIGEN DE LAS UNIDADES</span>
          <div class="kpi-value" style="font-size: 1.25rem;">{{ formatOrigen(sol.origenAsignacion) }}</div>
          <span class="kpi-sub">Solicitado desde {{ sol.hospitalOrigenId }}</span>
        </div>

        <div class="kpi-card">
          <span class="eyebrow">TIEMPO DE BÚSQUEDA</span>
          <div class="kpi-value">{{ sol.tiempoRespuestaMatchingMs }} ms</div>
          <span class="kpi-sub">Respuesta rápida</span>
        </div>
      </div>

      @if (sol.bancoAsignado; as banco) {
        <section class="surface-card" style="margin-bottom: 1.5rem;">
          <span class="eyebrow">BANCO QUE RESERVÓ LAS UNIDADES</span>
          <h2 class="sec-heading">{{ banco.nombre }}</h2>
          <dl class="meta-strip tabular">
            <div>
              <dt>Distancia aproximada</dt>
              <dd>{{ banco.distanciaKm }} km</dd>
            </div>
            <div>
              <dt>Tiempo estimado de traslado</dt>
              <dd>~{{ banco.tiempoEstimadoMin }} minutos</dd>
            </div>
            <div>
              <dt>Código del establecimiento</dt>
              <dd class="code-mono">{{ banco.bancoId }}</dd>
            </div>
          </dl>
        </section>
      }

      <!-- TABLA DE UNIDADES RESERVADAS -->
      <section class="surface-card" style="margin-bottom: 1.5rem;">
        <span class="eyebrow">UNIDADES RESERVADAS ({{ sol.unidadesAsignadas.length }})</span>
        @if (sol.unidadesAsignadas.length > 0) {
          <div class="table-wrap" style="margin-top: 0.9rem;">
            <table class="editorial-table tabular">
              <thead>
                <tr>
                  <th>Código de unidad</th>
                  <th>Código de barras</th>
                  <th>Historial</th>
                </tr>
              </thead>
              <tbody>
                @for (u of sol.unidadesAsignadas; track u.unidadId) {
                  <tr>
                    <td class="code-mono"><strong>{{ u.unidadId }}</strong></td>
                    <td class="code-mono">{{ u.codigoBarras }}</td>
                    <td>
                      @if (puedeVerTrazabilidad()) {
                        <a [routerLink]="['/trazabilidad', u.unidadId]" class="btn btn-ghost btn-sm">
                          Ver historial de la unidad
                        </a>
                      } @else {
                        <span>Unidad reservada</span>
                      }
                    </td>
                  </tr>
                }
              </tbody>
            </table>
          </div>
        } @else {
          <p style="margin-top: 0.75rem; color: var(--text-muted);">
            Esta solicitud aún no tiene bolsas reservadas ({{ formatEstado(sol.estadoSolicitud) }}).
          </p>
        }
      </section>

      @if (sol.motivoCancelacion) {
        <div style="margin-bottom: 1.5rem;">
          <app-inline-alert
            variant="info"
            title="Solicitud cancelada"
            [message]="'Motivo: ' + sol.motivoCancelacion"
          />
        </div>
      }

      <!-- ACCIONES DISPONIBLES -->
      @if (sol.estadoSolicitud === 'Reservada' || sol.estadoSolicitud === 'En Espera de Donantes') {
        <section class="surface-card actions-panel">
          <span class="eyebrow">SIGUIENTE PASO</span>

          <div class="actions-grid">
            @if (sol.estadoSolicitud === 'Reservada' && puedeConfirmar()) {
              <div class="action-block">
                <h3 class="action-title">Confirmar transfusión realizada</h3>
                <p class="action-desc">
                  Marca la solicitud como <strong>Atendida</strong> cuando las unidades hayan sido
                  entregadas y transfundidas al paciente.
                </p>
                <button
                  type="button"
                  class="btn btn-primary"
                  [disabled]="procesando()"
                  (click)="confirmarTransfusion()"
                >
                  {{ procesando() ? 'Guardando...' : 'Confirmar transfusión realizada' }}
                </button>
              </div>
            }

            @if (esMedico()) {
              <div class="action-block">
                <h3 class="action-title">Cancelar solicitud</h3>
                <p class="action-desc">
                  Anula este pedido y devuelve las bolsas reservadas al inventario disponible del banco.
                </p>
                <div class="cancel-inline">
                  <input
                    type="text"
                    class="input"
                    placeholder="Indica el motivo de cancelación..."
                    [(ngModel)]="motivoCancelacion"
                  />
                  <button
                    type="button"
                    class="btn btn-critical"
                    [disabled]="procesando()"
                    (click)="cancelarSolicitud()"
                  >
                    Cancelar solicitud
                  </button>
                </div>
              </div>
            }
          </div>
        </section>
      }
    }
  `,
  styles: [`
    .sec-heading { font-size: 1.35rem; font-weight: 900; margin: 0.35rem 0 1rem; }
    .meta-strip {
      display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 1rem; padding-top: 1rem; border-top: 1px solid var(--border-subtle);
    }
    .meta-strip dt { font-size: 0.78rem; font-weight: 700; color: var(--text-muted); }
    .meta-strip dd { font-size: 1rem; font-weight: 700; margin-top: 0.2rem; }
    .actions-panel { display: flex; flex-direction: column; gap: 1.15rem; }
    .actions-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 1.5rem; }
    .action-block { display: flex; flex-direction: column; gap: 0.65rem; }
    .action-title { font-size: 1.05rem; font-weight: 800; }
    .action-desc { font-size: 0.88rem; color: var(--text-secondary); }
    .cancel-inline { display: flex; flex-wrap: wrap; gap: 0.65rem; }
    .cancel-inline .input { flex: 1; min-width: 200px; }
  `],
})
export class EmergenciaDetalleComponent implements OnInit {
  private readonly emergenciasService = inject(EmergenciasService);
  private readonly authService = inject(AuthService);

  readonly solicitudId = input.required<string>();

  protected readonly solicitud = signal<EmergenciaMatchingResponseDTO | null>(null);
  protected readonly procesando = signal(false);
  protected readonly apiError = signal<ExtractedError | null>(null);
  protected readonly accionExitosa = signal<string>('');
  protected motivoCancelacion = '';

  protected readonly formatEstado = formatEstadoSolicitud;
  protected readonly formatOrigen = formatOrigenAsignacion;
  protected readonly formatUrg = formatUrgencia;

  protected readonly esMedico = computed(() => this.authService.currentRole() === 'MEDICO');
  protected readonly puedeConfirmar = computed(() =>
    this.authService.hasRole(['MEDICO', 'PERSONAL_BANCO'])
  );
  protected readonly puedeVerTrazabilidad = computed(() =>
    this.authService.hasRole(['COORDINADOR', 'PERSONAL_BANCO'])
  );

  ngOnInit(): void {
    this.cargarDetalle();
  }

  cargarDetalle(): void {
    this.apiError.set(null);
    this.emergenciasService.obtenerSolicitud(this.solicitudId()).subscribe({
      next: (data) => this.solicitud.set(data),
      error: (err) => this.apiError.set(extractApiError(err)),
    });
  }

  confirmarTransfusion(): void {
    const sol = this.solicitud();
    if (!sol) return;
    this.procesando.set(true);
    this.apiError.set(null);
    this.accionExitosa.set('');

    this.emergenciasService.confirmarTransfusion(sol.solicitudId).subscribe({
      next: (updated) => {
        this.procesando.set(false);
        this.solicitud.set(updated);
        this.accionExitosa.set(
          'Transfusión confirmada. Las unidades quedaron registradas como Transfundidas.'
        );
      },
      error: (err) => {
        this.procesando.set(false);
        this.apiError.set(extractApiError(err));
      },
    });
  }

  cancelarSolicitud(): void {
    const sol = this.solicitud();
    if (!sol) return;
    this.procesando.set(true);
    this.apiError.set(null);
    this.accionExitosa.set('');

    const motivo = this.motivoCancelacion.trim() || 'Cancelada por el médico solicitante';
    this.emergenciasService.cancelarSolicitud(sol.solicitudId, { motivo }).subscribe({
      next: (updated) => {
        this.procesando.set(false);
        this.solicitud.set(updated);
        this.accionExitosa.set(
          'Solicitud cancelada. Las unidades reservadas volvieron a estar disponibles.'
        );
      },
      error: (err) => {
        this.procesando.set(false);
        this.apiError.set(extractApiError(err));
      },
    });
  }
}
