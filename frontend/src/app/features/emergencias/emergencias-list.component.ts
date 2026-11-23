import { ChangeDetectionStrategy, Component, OnInit, computed, inject, signal } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { EmergenciaMatchingResponseDTO, EstadoSolicitud } from '../../core/models/emergencia.model';
import { AuthService } from '../../core/services/auth.service';
import { EmergenciasService } from '../../core/services/emergencias.service';
import { ExtractedError, extractApiError } from '../../core/utils/error.util';
import {
  formatEstadoSolicitud,
  formatFechaCorta,
  formatHospital,
  formatOrigenAsignacion,
  formatUrgencia,
} from '../../core/utils/label.util';
import { EmptyStateComponent } from '../../shared/ui/empty-state/empty-state.component';
import { PageHeaderComponent } from '../../shared/ui/page-header/page-header.component';
import { IndicatorTone, StatusIndicatorComponent } from '../../shared/ui/status-indicator/status-indicator.component';

interface FiltroEstadoOption {
  label: string;
  value: string;
}

const FILTROS_ESTADO: readonly FiltroEstadoOption[] = [
  { label: 'Todas', value: '' },
  { label: 'Reservadas', value: 'Reservada' },
  { label: 'Buscando donantes', value: 'En Espera de Donantes' },
  { label: 'Atendidas', value: 'Atendida' },
  { label: 'Canceladas', value: 'Cancelada' },
] as const;

const SKELETON_ROWS = [1, 2, 3, 4, 5] as const;

@Component({
  selector: 'app-emergencias-list',
  standalone: true,
  imports: [
    RouterLink,
    PageHeaderComponent,
    StatusIndicatorComponent,
    EmptyStateComponent,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <app-page-header title="Emergencias">
      @if (esMedico()) {
        <a routerLink="/emergencias/nueva" class="btn btn-primary">
          Nueva solicitud
        </a>
      }
    </app-page-header>

    <div class="emerg-controls">
      <div class="pill-tablist" role="tablist" aria-label="Filtrar solicitudes por estado">
        @for (opt of filtrosEstado; track opt.value) {
          <button
            type="button"
            role="tab"
            class="filter-pill"
            [class.active]="estadoFiltro() === opt.value"
            [attr.aria-selected]="estadoFiltro() === opt.value"
            (click)="seleccionarFiltro(opt.value)"
          >
            {{ opt.label }}
          </button>
        }
      </div>

      <button
        type="button"
        class="btn-refresh"
        [disabled]="loading()"
        (click)="cargarSolicitudes()"
      >
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          aria-hidden="true"
          [class.spin-icon]="loading()"
        >
          <path
            d="M20 12A8 8 0 1 1 17.66 6.34M20 4V9H15"
            stroke="currentColor"
            stroke-width="2.1"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
        <span>{{ loading() ? 'Actualizando...' : 'Actualizar' }}</span>
      </button>
    </div>

    @if (apiError(); as err) {
      <div class="inline-error-row">
        <p class="field-error" role="alert">{{ formatErrorText(err) }}</p>
        <button type="button" class="btn-retry" (click)="cargarSolicitudes()">
          Reintentar
        </button>
      </div>
    }

    @if (loading()) {
      <div class="emerg-card" aria-busy="true" aria-label="Cargando emergencias">
        <table class="emerg-table" aria-hidden="true">
          <thead>
            <tr>
              <th>Grupo</th>
              <th>Origen / Asignación</th>
              <th>Prioridad</th>
              <th>Estado</th>
              <th>Fecha</th>
              <th>Acción</th>
            </tr>
          </thead>
          <tbody>
            @for (row of skeletonRows; track row) {
              <tr>
                <td>
                  <div class="skel-stack">
                    <span class="skeleton" style="width: 46px; height: 24px;"></span>
                    <span class="skeleton" style="width: 36px; height: 14px;"></span>
                  </div>
                </td>
                <td>
                  <div class="skel-stack">
                    <span class="skeleton" style="width: 170px; height: 16px;"></span>
                    <span class="skeleton" style="width: 120px; height: 13px;"></span>
                  </div>
                </td>
                <td>
                  <span class="skeleton" style="width: 110px; height: 16px;"></span>
                </td>
                <td>
                  <div class="skel-row-inline">
                    <span class="skeleton" style="width: 10px; height: 10px; flex-shrink: 0;"></span>
                    <span class="skeleton" style="width: 95px; height: 16px;"></span>
                  </div>
                </td>
                <td>
                  <span class="skeleton" style="width: 110px; height: 15px;"></span>
                </td>
                <td>
                  <span class="skeleton" style="width: 48px; height: 16px;"></span>
                </td>
              </tr>
            }
          </tbody>
        </table>

        <div class="emerg-mobile-list" aria-hidden="true">
          @for (row of skeletonRows; track row) {
            <div class="emerg-mobile-card">
              <div class="emerg-mobile-top">
                <div class="skel-row-inline">
                  <span class="skeleton" style="width: 46px; height: 24px;"></span>
                  <span class="skeleton" style="width: 36px; height: 14px;"></span>
                </div>
                <div class="skel-row-inline">
                  <span class="skeleton" style="width: 10px; height: 10px;"></span>
                  <span class="skeleton" style="width: 95px; height: 16px;"></span>
                </div>
              </div>
              <div class="skel-stack">
                <span class="skeleton" style="width: 170px; height: 16px;"></span>
                <span class="skeleton" style="width: 120px; height: 13px;"></span>
              </div>
              <div class="emerg-mobile-bottom">
                <span class="skeleton" style="width: 110px; height: 15px;"></span>
                <span class="skeleton" style="width: 48px; height: 16px;"></span>
              </div>
            </div>
          }
        </div>
      </div>
    } @else if (solicitudes().length === 0) {
      <app-empty-state
        eyebrow="SIN RESULTADOS"
        [title]="estadoFiltro() ? 'No hay solicitudes en este estado' : 'Aún no hay solicitudes registradas'"
        description="Cuando un médico solicite unidades para una emergencia, aparecerán aquí en tiempo real."
      >
        @if (estadoFiltro()) {
          <button type="button" class="btn btn-secondary" (click)="seleccionarFiltro('')">
            Ver todas
          </button>
        } @else if (esMedico()) {
          <a routerLink="/emergencias/nueva" class="btn btn-primary">Nueva solicitud</a>
        }
      </app-empty-state>
    } @else {
      <div class="emerg-card">
        <table class="emerg-table tabular">
          <thead>
            <tr>
              <th>Grupo</th>
              <th>Origen / Asignación</th>
              <th>Prioridad</th>
              <th>Estado</th>
              <th>Fecha</th>
              <th>Acción</th>
            </tr>
          </thead>
          <tbody>
            @for (sol of solicitudes(); track sol.solicitudId) {
              <tr class="clickable-row" (click)="verDetalle(sol.solicitudId)">
                <td>
                  <div class="blood-group-cell">
                    <span class="blood-main">{{ sol.tipoSangreReq || 'O' }}{{ sol.factorRhReq || '+' }}</span>
                    <span class="blood-qty">{{ sol.cantidadReq ?? 1 }} u.</span>
                  </div>
                </td>
                <td>
                  <div class="cell-primary">{{ formatHospitalLabel(sol.hospitalOrigenId) }}</div>
                  <div class="cell-sub">{{ formatAsignacion(sol) }}</div>
                </td>
                <td>
                  <span class="priority-label" [class.priority-vital]="sol.urgencia === 'ALTA_VITAL'">
                    {{ formatUrgenciaLabel(sol.urgencia) }}
                  </span>
                </td>
                <td>
                  <app-status-indicator
                    [tone]="toneForEstado(sol.estadoSolicitud)"
                    [label]="formatEstado(sol.estadoSolicitud)"
                  />
                </td>
                <td class="date-cell">{{ formatFecha(sol.fechaHora) }}</td>
                <td>
                  <a
                    [routerLink]="['/emergencias', sol.solicitudId]"
                    class="row-ver-link"
                    (click)="$event.stopPropagation()"
                  >
                    <span>Ver</span>
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                      <path d="M9 6L15 12L9 18" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
                    </svg>
                  </a>
                </td>
              </tr>
            }
          </tbody>
        </table>

        <div class="emerg-mobile-list tabular">
          @for (sol of solicitudes(); track sol.solicitudId) {
            <article class="emerg-mobile-card" (click)="verDetalle(sol.solicitudId)">
              <div class="emerg-mobile-top">
                <div class="skel-row-inline">
                  <span class="blood-main">{{ sol.tipoSangreReq || 'O' }}{{ sol.factorRhReq || '+' }}</span>
                  <span class="blood-qty">· {{ sol.cantidadReq ?? 1 }} u.</span>
                </div>
                <app-status-indicator
                  [tone]="toneForEstado(sol.estadoSolicitud)"
                  [label]="formatEstado(sol.estadoSolicitud)"
                />
              </div>

              <div>
                <div class="cell-primary">{{ formatHospitalLabel(sol.hospitalOrigenId) }}</div>
                <div class="cell-sub">{{ formatAsignacion(sol) }}</div>
              </div>

              <div class="emerg-mobile-bottom">
                <span class="priority-label" [class.priority-vital]="sol.urgencia === 'ALTA_VITAL'">
                  {{ formatUrgenciaLabel(sol.urgencia) }}
                </span>
                <span class="date-cell">{{ formatFecha(sol.fechaHora) }}</span>
              </div>

              <div class="emerg-mobile-bottom">
                <span></span>
                <a
                  [routerLink]="['/emergencias', sol.solicitudId]"
                  class="row-ver-link"
                  (click)="$event.stopPropagation()"
                >
                  <span>Ver</span>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                    <path d="M9 6L15 12L9 18" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                </a>
              </div>
            </article>
          }
        </div>
      </div>
    }
  `,
})
export class EmergenciasListComponent implements OnInit {
  private readonly emergenciasService = inject(EmergenciasService);
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);

  protected readonly filtrosEstado = FILTROS_ESTADO;
  protected readonly skeletonRows = SKELETON_ROWS;
  protected readonly solicitudes = signal<EmergenciaMatchingResponseDTO[]>([]);
  protected readonly loading = signal(false);
  protected readonly apiError = signal<ExtractedError | null>(null);
  protected readonly estadoFiltro = signal('');
  protected readonly esMedico = computed(() => this.authService.currentRole() === 'MEDICO');

  protected readonly formatEstado = formatEstadoSolicitud;
  protected readonly formatHospitalLabel = formatHospital;
  protected readonly formatUrgenciaLabel = formatUrgencia;
  protected readonly formatFecha = formatFechaCorta;

  ngOnInit(): void {
    this.cargarSolicitudes();
  }

  seleccionarFiltro(estado: string): void {
    if (this.estadoFiltro() === estado) return;
    this.estadoFiltro.set(estado);
    this.cargarSolicitudes();
  }

  cargarSolicitudes(): void {
    this.loading.set(true);
    this.apiError.set(null);

    const filtro = this.estadoFiltro() || undefined;
    this.emergenciasService.listarSolicitudes(filtro).subscribe({
      next: (list) => {
        this.loading.set(false);
        this.solicitudes.set(list);
      },
      error: (err) => {
        this.loading.set(false);
        this.apiError.set(extractApiError(err));
      },
    });
  }

  verDetalle(solicitudId: string): void {
    void this.router.navigate(['/emergencias', solicitudId]);
  }

  formatAsignacion(sol: EmergenciaMatchingResponseDTO): string {
    if (sol.bancoAsignado) {
      const nombreCorto = formatHospital(sol.bancoAsignado.bancoId);
      return nombreCorto !== sol.bancoAsignado.bancoId ? nombreCorto : sol.bancoAsignado.nombre;
    }
    if (sol.estadoSolicitud === 'En Espera de Donantes' || sol.origenAsignacion === 'NINGUNO') {
      return 'Buscando en red';
    }
    return formatOrigenAsignacion(sol.origenAsignacion);
  }

  formatErrorText(err: ExtractedError): string {
    const raw = (err.mensaje || err.detalle || '').trim();
    if (!raw) return 'No pudimos cargar las emergencias. Intenta nuevamente.';
    const cap = raw.charAt(0).toUpperCase() + raw.slice(1);
    return /[.!?]$/.test(cap) ? cap : `${cap}.`;
  }

  toneForEstado(estado: EstadoSolicitud): IndicatorTone {
    switch (estado) {
      case 'Atendida':
        return 'ok';
      case 'Reservada':
        return 'ink';
      case 'En Espera de Donantes':
        return 'critical';
      case 'Cancelada':
        return 'neutral';
      default:
        return 'warning';
    }
  }
}
