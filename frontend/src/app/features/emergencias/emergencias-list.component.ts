import { ChangeDetectionStrategy, Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { EmergenciaMatchingResponseDTO, EstadoSolicitud } from '../../core/models/emergencia.model';
import { AuthService } from '../../core/services/auth.service';
import { EmergenciasService } from '../../core/services/emergencias.service';
import { ExtractedError, extractApiError } from '../../core/utils/error.util';
import { formatEstadoSolicitud, formatOrigenAsignacion } from '../../core/utils/label.util';
import { EmptyStateComponent } from '../../shared/ui/empty-state/empty-state.component';
import { InlineAlertComponent } from '../../shared/ui/inline-alert/inline-alert.component';
import { PageHeaderComponent } from '../../shared/ui/page-header/page-header.component';
import { IndicatorTone, StatusIndicatorComponent } from '../../shared/ui/status-indicator/status-indicator.component';

@Component({
  selector: 'app-emergencias-list',
  standalone: true,
  imports: [
    FormsModule,
    RouterLink,
    PageHeaderComponent,
    InlineAlertComponent,
    StatusIndicatorComponent,
    EmptyStateComponent,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <app-page-header
      eyebrow="ATENCIÓN DE EMERGENCIAS"
      title="Solicitudes de emergencia"
      subtitle="Seguimiento en tiempo real de unidades reservadas, traslados desde bancos cercanos y tiempos de respuesta."
    >
      @if (esMedico()) {
        <a routerLink="/emergencias/nueva" class="btn btn-primary">
          Pedir unidades
        </a>
      }
    </app-page-header>

    <div class="filter-bar">
      <div class="field">
        <label class="field-label" for="estadoFiltro">Filtrar por estado</label>
        <select
          id="estadoFiltro"
          class="select"
          [(ngModel)]="estadoFiltro"
          (ngModelChange)="cargarSolicitudes()"
        >
          <option value="">Todos los estados</option>
          <option value="Reservada">Reservada</option>
          <option value="Atendida">Atendida</option>
          <option value="En Espera de Donantes">Buscando donantes</option>
          <option value="Cancelada">Cancelada</option>
        </select>
      </div>
      <div>
        <button type="button" class="btn btn-ghost" (click)="cargarSolicitudes()">
          Actualizar listado
        </button>
      </div>
    </div>

    @if (apiError(); as err) {
      <div style="margin-bottom: 1.5rem;">
        <app-inline-alert variant="error" [title]="err.mensaje" [message]="err.detalle" />
      </div>
    }

    @if (solicitudes().length === 0 && !loading()) {
      <app-empty-state
        eyebrow="EMERGENCIAS"
        title="Aún no hay solicitudes registradas"
        description="Cuando un médico solicite unidades para una emergencia, aparecerán aquí junto con el banco que las reservó."
      >
        @if (esMedico()) {
          <a routerLink="/emergencias/nueva" class="btn btn-primary">Solicitar unidades</a>
        }
      </app-empty-state>
    } @else {
      <div class="table-wrap">
        <table class="editorial-table tabular">
          <thead>
            <tr>
              <th>Solicitud</th>
              <th>Grupo y cantidad</th>
              <th>Hospital solicitante</th>
              <th>Estado</th>
              <th>Banco asignado</th>
              <th>Tiempo</th>
              <th>Detalle</th>
            </tr>
          </thead>
          <tbody>
            @for (sol of solicitudes(); track sol.solicitudId) {
              <tr>
                <td class="code-mono">
                  <strong>{{ sol.solicitudId }}</strong>
                </td>
                <td>
                  <strong>{{ sol.tipoSangreReq }}{{ sol.factorRhReq }}</strong>
                  <span> · {{ sol.cantidadReq }} u.</span>
                </td>
                <td>{{ sol.hospitalOrigenId }}</td>
                <td>
                  <app-status-indicator
                    [tone]="toneForEstado(sol.estadoSolicitud)"
                    [label]="formatEstado(sol.estadoSolicitud)"
                  />
                </td>
                <td>
                  @if (sol.bancoAsignado; as b) {
                    <div>{{ b.nombre }}</div>
                    <div class="row-sub">
                      {{ formatOrigen(sol.origenAsignacion) }} · {{ b.distanciaKm }} km · ~{{ b.tiempoEstimadoMin }} min
                    </div>
                  } @else {
                    <span class="row-sub">{{ formatOrigen(sol.origenAsignacion) }}</span>
                  }
                </td>
                <td>{{ sol.tiempoRespuestaMatchingMs }} ms</td>
                <td>
                  <a [routerLink]="['/emergencias', sol.solicitudId]" class="btn btn-ghost btn-sm">
                    Ver detalle
                  </a>
                </td>
              </tr>
            }
          </tbody>
        </table>
      </div>
    }
  `,
  styles: [`
    .row-sub {
      font-size: 0.8rem;
      font-weight: 500;
      color: var(--text-muted);
    }
  `],
})
export class EmergenciasListComponent implements OnInit {
  private readonly emergenciasService = inject(EmergenciasService);
  private readonly authService = inject(AuthService);

  protected readonly solicitudes = signal<EmergenciaMatchingResponseDTO[]>([]);
  protected readonly loading = signal(false);
  protected readonly apiError = signal<ExtractedError | null>(null);
  protected readonly esMedico = computed(() => this.authService.currentRole() === 'MEDICO');
  protected readonly formatEstado = formatEstadoSolicitud;
  protected readonly formatOrigen = formatOrigenAsignacion;
  protected estadoFiltro = '';

  ngOnInit(): void {
    this.cargarSolicitudes();
  }

  cargarSolicitudes(): void {
    this.loading.set(true);
    this.apiError.set(null);

    this.emergenciasService.listarSolicitudes(this.estadoFiltro || undefined).subscribe({
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
