import { ChangeDetectionStrategy, Component, OnInit, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { FACTORES_RH, GRUPOS_SANGUINEOS } from '../../core/config/api.config';
import { DonanteResponseDTO } from '../../core/models/donante.model';
import { DonantesService } from '../../core/services/donantes.service';
import { ExtractedError, extractApiError } from '../../core/utils/error.util';
import { EmptyStateComponent } from '../../shared/ui/empty-state/empty-state.component';
import { InlineAlertComponent } from '../../shared/ui/inline-alert/inline-alert.component';
import { PageHeaderComponent } from '../../shared/ui/page-header/page-header.component';
import { StatusIndicatorComponent } from '../../shared/ui/status-indicator/status-indicator.component';

@Component({
  selector: 'app-donantes-list',
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
      eyebrow="COMUNIDAD DE DONANTES"
      title="Donantes voluntarios registrados"
      subtitle="Personas dispuestas a donar en la región. Por privacidad, su DNI y teléfono se muestran parcialmente ocultos."
    >
      <a routerLink="/donantes/registro" class="btn btn-primary">
        Registrar nuevo donante
      </a>
    </app-page-header>

    <div class="filter-bar">
      <div class="field">
        <label class="field-label" for="dTipo">Grupo sanguíneo</label>
        <select id="dTipo" class="select" [(ngModel)]="tipoSangre" (ngModelChange)="cargarDonantes()">
          <option value="">Todos los grupos</option>
          @for (g of grupos; track g) {
            <option [value]="g">Grupo {{ g }}</option>
          }
        </select>
      </div>

      <div class="field">
        <label class="field-label" for="dRh">Factor Rh</label>
        <select id="dRh" class="select" [(ngModel)]="factorRh" (ngModelChange)="cargarDonantes()">
          <option value="">Ambos (+ / -)</option>
          @for (rh of factoresRh; track rh) {
            <option [value]="rh">Rh {{ rh }}</option>
          }
        </select>
      </div>
    </div>

    @if (apiError(); as err) {
      <div style="margin-bottom: 1.5rem;">
        <app-inline-alert variant="error" [title]="err.mensaje" [message]="err.detalle" />
      </div>
    }

    @if (mensajeBaja(); as msg) {
      <div style="margin-bottom: 1.5rem;">
        <app-inline-alert variant="info" title="Registro actualizado" [message]="msg" />
      </div>
    }

    @if (donantes().length === 0) {
      <app-empty-state
        eyebrow="DONANTES"
        title="No hay donantes registrados para este grupo"
        description="Cualquier ciudadano mayor de edad puede registrarse en menos de 3 minutos desde el portal público."
      />
    } @else {
      <div class="table-wrap">
        <table class="editorial-table tabular">
          <thead>
            <tr>
              <th>Código</th>
              <th>Nombre completo</th>
              <th>DNI protegido</th>
              <th>Grupo</th>
              <th>Contacto</th>
              <th>Hospital preferido</th>
              <th>Disponibilidad</th>
              <th>Opción</th>
            </tr>
          </thead>
          <tbody>
            @for (d of donantes(); track d.donanteId) {
              <tr>
                <td class="code-mono"><strong>{{ d.donanteId }}</strong></td>
                <td>{{ d.nombreCompleto }}</td>
                <td class="code-mono">{{ d.dni }}</td>
                <td><strong>{{ d.tipoSangre }}{{ d.factorRh }}</strong></td>
                <td>
                  <div class="code-mono">{{ d.celular }}</div>
                  <div style="font-size: 0.78rem; color: var(--text-muted);">{{ d.correo }}</div>
                </td>
                <td>{{ d.bancoPreferidoId }}</td>
                <td>
                  <app-status-indicator
                    [tone]="d.habilitadoConvocatorias ? 'ok' : 'neutral'"
                    [label]="d.habilitadoConvocatorias ? 'Disponible para avisos' : 'Inactivo'"
                  />
                </td>
                <td>
                  @if (confirmandoBajaId() === d.donanteId) {
                    <div class="inline-confirm">
                      <button type="button" class="btn btn-critical btn-sm" (click)="confirmarBaja(d.donanteId)">
                        Confirmar baja
                      </button>
                      <button type="button" class="btn btn-ghost btn-sm" (click)="confirmandoBajaId.set(null)">
                        Cancelar
                      </button>
                    </div>
                  } @else {
                    <button
                      type="button"
                      class="btn btn-ghost btn-sm"
                      (click)="confirmandoBajaId.set(d.donanteId)"
                    >
                      Dar de baja
                    </button>
                  }
                </td>
              </tr>
            }
          </tbody>
        </table>
      </div>
    }
  `,
  styles: [`
    .inline-confirm { display: flex; gap: 0.45rem; }
  `],
})
export class DonantesListComponent implements OnInit {
  private readonly donantesService = inject(DonantesService);

  protected readonly grupos = GRUPOS_SANGUINEOS;
  protected readonly factoresRh = FACTORES_RH;

  protected readonly donantes = signal<DonanteResponseDTO[]>([]);
  protected readonly apiError = signal<ExtractedError | null>(null);
  protected readonly mensajeBaja = signal<string>('');
  protected readonly confirmandoBajaId = signal<string | null>(null);

  protected tipoSangre = '';
  protected factorRh = '';

  ngOnInit(): void {
    this.cargarDonantes();
  }

  cargarDonantes(): void {
    this.apiError.set(null);
    this.donantesService
      .listarDonantes(this.tipoSangre || undefined, this.factorRh || undefined)
      .subscribe({
        next: (list) => this.donantes.set(list),
        error: (err) => this.apiError.set(extractApiError(err)),
      });
  }

  confirmarBaja(donanteId: string): void {
    this.apiError.set(null);
    this.donantesService.darDeBaja(donanteId).subscribe({
      next: (res) => {
        this.confirmandoBajaId.set(null);
        this.mensajeBaja.set(`${res.mensaje} (${res.donanteId}).`);
        this.cargarDonantes();
      },
      error: (err) => this.apiError.set(extractApiError(err)),
    });
  }
}
