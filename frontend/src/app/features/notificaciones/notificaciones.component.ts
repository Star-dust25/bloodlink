import { ChangeDetectionStrategy, Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import {
  Notificacion,
  VerificacionCaducidadResponse,
} from '../../core/models/notificacion.model';
import { AuthService } from '../../core/services/auth.service';
import { NotificacionesService } from '../../core/services/notificaciones.service';
import { ExtractedError, extractApiError } from '../../core/utils/error.util';
import {
  formatCanalNotificacion,
  formatEventoNotificacion,
} from '../../core/utils/label.util';
import { EmptyStateComponent } from '../../shared/ui/empty-state/empty-state.component';
import { InlineAlertComponent } from '../../shared/ui/inline-alert/inline-alert.component';
import { PageHeaderComponent } from '../../shared/ui/page-header/page-header.component';
import { StatusIndicatorComponent } from '../../shared/ui/status-indicator/status-indicator.component';

@Component({
  selector: 'app-notificaciones',
  standalone: true,
  imports: [
    FormsModule,
    PageHeaderComponent,
    InlineAlertComponent,
    StatusIndicatorComponent,
    EmptyStateComponent,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <app-page-header
      eyebrow="COMUNICACIÓN Y PREVENCIÓN"
      title="Alertas y avisos"
      subtitle="Registro de avisos enviados al personal de salud y convocatorias enviadas a donantes voluntarios."
    >
      @if (puedeEjecutarJob()) {
        <button
          type="button"
          class="btn btn-primary"
          [disabled]="ejecutandoJob()"
          (click)="verificarCaducidad()"
        >
          {{ ejecutandoJob() ? 'Revisando...' : 'Revisar unidades por vencer (próximos 7 días)' }}
        </button>
      }
    </app-page-header>

    @if (apiError(); as err) {
      <div style="margin-bottom: 1.5rem;">
        <app-inline-alert variant="error" [title]="err.mensaje" [message]="err.detalle" />
      </div>
    }

    @if (resultadoJob(); as job) {
      <div style="margin-bottom: 1.5rem;">
        <app-inline-alert
          [variant]="job.totalDetectadas > 0 ? 'warning' : 'success'"
          [title]="'Revisión completada: ' + job.totalDetectadas + ' unidades vencen en los próximos 7 días'"
          [message]="'Se emitieron ' + job.alertasPublicadas + ' alertas preventivas para priorizar su uso antes del vencimiento.'"
        />
      </div>
    }

    <div class="filter-bar">
      <div class="field">
        <label class="field-label" for="canalSelect">Medio de envío</label>
        <select
          id="canalSelect"
          class="select"
          [(ngModel)]="canalFiltro"
          (ngModelChange)="cargarNotificaciones()"
        >
          <option value="">Todos los medios</option>
          <option value="ALERTA_PANEL">Aviso en sistema</option>
          <option value="SMS">Mensaje SMS (convocatorias)</option>
          <option value="EMAIL">Correo electrónico (convocatorias)</option>
        </select>
      </div>
      <div>
        <button type="button" class="btn btn-ghost" (click)="cargarNotificaciones()">
          Actualizar avisos
        </button>
      </div>
    </div>

    @if (notificaciones().length === 0) {
      <app-empty-state
        eyebrow="ALERTAS"
        title="No hay avisos para este filtro"
        description="Aquí aparecerán las reservas de emergencia, alertas de vencimiento y convocatorias a donantes."
      />
    } @else {
      <div class="table-wrap">
        <table class="editorial-table tabular">
          <thead>
            <tr>
              <th>Tipo de aviso</th>
              <th>Medio</th>
              <th>Destinatario</th>
              <th>Detalle del mensaje</th>
              <th>Fecha</th>
            </tr>
          </thead>
          <tbody>
            @for (n of notificaciones(); track n.id) {
              <tr>
                <td>
                  <app-status-indicator
                    [tone]="n.evento === 'StockCriticoDetectado' ? 'critical' : n.evento === 'UnidadProximaACaducar' ? 'warning' : 'ink'"
                    [label]="formatEvento(n.evento)"
                  />
                </td>
                <td>{{ formatCanal(n.canal) }}</td>
                <td>{{ n.destinatario }}</td>
                <td>{{ n.mensaje }}</td>
                <td class="code-mono" style="font-size: 0.78rem; color: var(--text-muted);">
                  {{ n.timestamp }}
                </td>
              </tr>
            }
          </tbody>
        </table>
      </div>
    }
  `,
})
export class NotificacionesComponent implements OnInit {
  private readonly notificacionesService = inject(NotificacionesService);
  private readonly authService = inject(AuthService);

  protected readonly notificaciones = signal<Notificacion[]>([]);
  protected readonly resultadoJob = signal<VerificacionCaducidadResponse | null>(null);
  protected readonly ejecutandoJob = signal(false);
  protected readonly apiError = signal<ExtractedError | null>(null);
  protected canalFiltro = '';

  protected readonly formatEvento = formatEventoNotificacion;
  protected readonly formatCanal = formatCanalNotificacion;

  protected readonly puedeEjecutarJob = computed(() =>
    this.authService.hasRole(['PERSONAL_BANCO', 'COORDINADOR'])
  );

  ngOnInit(): void {
    this.cargarNotificaciones();
  }

  cargarNotificaciones(): void {
    this.apiError.set(null);
    this.notificacionesService
      .listarNotificaciones(this.canalFiltro || undefined, 80)
      .subscribe({
        next: (list) => this.notificaciones.set(list),
        error: (err) => this.apiError.set(extractApiError(err)),
      });
  }

  verificarCaducidad(): void {
    this.ejecutandoJob.set(true);
    this.apiError.set(null);

    this.notificacionesService.ejecutarVerificacionCaducidad().subscribe({
      next: (res) => {
        this.ejecutandoJob.set(false);
        this.resultadoJob.set(res);
        setTimeout(() => this.cargarNotificaciones(), 150);
      },
      error: (err) => {
        this.ejecutandoJob.set(false);
        this.apiError.set(extractApiError(err));
      },
    });
  }
}
