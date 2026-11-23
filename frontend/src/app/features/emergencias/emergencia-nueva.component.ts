import { HttpErrorResponse } from '@angular/common/http';
import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import {
  BANCOS_PIURA_CATALOGO,
  FACTORES_RH,
  GRUPOS_SANGUINEOS,
} from '../../core/config/api.config';
import {
  CrearEmergenciaRequest,
  EmergenciaSinStockConflictDTO,
} from '../../core/models/emergencia.model';
import { FactorRh, TipoSangre } from '../../core/models/inventario.model';
import { AuthService } from '../../core/services/auth.service';
import { EmergenciasService } from '../../core/services/emergencias.service';
import { ExtractedError, extractApiError } from '../../core/utils/error.util';
import { InlineAlertComponent } from '../../shared/ui/inline-alert/inline-alert.component';
import { PageHeaderComponent } from '../../shared/ui/page-header/page-header.component';
import { StatusIndicatorComponent } from '../../shared/ui/status-indicator/status-indicator.component';

@Component({
  selector: 'app-emergencia-nueva',
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
    <div class="narrow-workspace">
      <app-page-header
        eyebrow="SOLICITUD URGENTE"
        title="Solicitar unidades de emergencia"
        subtitle="Buscamos primero en tu propio hospital y, si no hay stock suficiente, reservamos automáticamente en el banco más cercano."
      >
        <a routerLink="/emergencias" class="btn btn-ghost">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M15 18L9 12L15 6" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
          <span>Volver al listado</span>
        </a>
      </app-page-header>

      @if (apiError(); as err) {
        <div style="margin-bottom: 1.5rem;">
          <app-inline-alert variant="error" [title]="err.mensaje" [message]="err.detalle" />
        </div>
      }

      <!-- ESCENARIO SIN STOCK EN LA RED: CONVOCATORIA DE DONANTES -->
      @if (conflictoSinStock(); as conf) {
        <section class="surface-card-strong conflict-panel">
          <div class="conflict-head">
            <app-status-indicator
              tone="critical"
              label="Buscando donantes"
              [detail]="'Búsqueda completada en ' + conf.tiempoRespuestaMatchingMs + ' ms'"
            />
            <h2 class="conflict-id tabular">{{ conf.solicitudId }}</h2>
          </div>

          <app-inline-alert
            variant="error"
            title="No encontramos unidades disponibles en los bancos cercanos"
            message="Revisamos tu hospital y todos los bancos conectados en la región, pero actualmente no cuentan con las unidades solicitadas."
          />

          <div class="convocatoria-box">
            <span class="eyebrow">CONVOCATORIA AUTOMÁTICA ACTIVADA</span>
            <p class="convocatoria-copy">
              Tu solicitud quedó registrada en estado <strong>Buscando donantes</strong> y ya
              enviamos avisos por <strong>mensaje SMS y correo electrónico</strong> a los donantes
              voluntarios compatibles con el grupo
              <strong class="tabular">{{ tipoSangreReq }}{{ factorRhReq }}</strong>.
            </p>
          </div>

          <div class="conflict-actions">
            <a [routerLink]="['/emergencias', conf.solicitudId]" class="btn btn-primary">
              Ver ficha de solicitud
            </a>
            <a routerLink="/notificaciones" class="btn btn-secondary">
              Ver avisos enviados
            </a>
            <button type="button" class="btn btn-ghost" (click)="conflictoSinStock.set(null)">
              Intentar otra búsqueda
            </button>
          </div>
        </section>
      } @else {
        <!-- FORMULARIO ENFOCADO -->
        <section class="surface-card">
          <form class="form-stack" (ngSubmit)="ejecutarMatching()">
            <div class="form-row">
              <div class="field">
                <label class="field-label" for="hospOrigen">Hospital solicitante</label>
                <select
                  id="hospOrigen"
                  name="hospitalOrigenId"
                  class="select"
                  [(ngModel)]="hospitalOrigenId"
                  (ngModelChange)="onHospitalChange($event)"
                >
                  @for (b of bancos; track b.id) {
                    <option [value]="b.id">{{ b.nombre }}</option>
                  }
                </select>
              </div>

              <div class="field">
                <label class="field-label" for="coordOrigen">Ubicación del hospital</label>
                <input
                  id="coordOrigen"
                  name="coordenadasOrigen"
                  type="text"
                  class="input code-mono tabular"
                  [(ngModel)]="coordenadasOrigen"
                  required
                />
              </div>
            </div>

            <div class="form-row">
              <div class="field">
                <span class="field-label">Grupo sanguíneo requerido</span>
                <div class="selector-grid">
                  @for (g of grupos; track g) {
                    <button
                      type="button"
                      class="selector-btn tabular"
                      [class.selected]="tipoSangreReq === g"
                      (click)="tipoSangreReq = g"
                    >
                      {{ g }}
                    </button>
                  }
                </div>
              </div>

              <div class="field">
                <span class="field-label">Factor Rh requerido</span>
                <div class="selector-grid">
                  @for (rh of factoresRh; track rh) {
                    <button
                      type="button"
                      class="selector-btn tabular"
                      [class.selected]="factorRhReq === rh"
                      (click)="factorRhReq = rh"
                    >
                      {{ rh }}
                    </button>
                  }
                </div>
              </div>
            </div>

            <div class="form-row">
              <div class="field">
                <label class="field-label" for="cantidadReq">Cantidad de unidades necesarias</label>
                <input
                  id="cantidadReq"
                  name="cantidadReq"
                  type="number"
                  min="1"
                  max="20"
                  class="input tabular"
                  required
                  [(ngModel)]="cantidadReq"
                />
                <span class="field-hint">
                  Tip rápido: pide <strong>2 unidades O+</strong> desde <strong>Hospital Santa Rosa</strong> para ubicar y reservar en el banco cercano.
                </span>
              </div>

              <div class="field">
                <span class="field-label">Nivel de urgencia</span>
                <label class="checkbox-row" style="padding-top: 0.65rem;">
                  <input
                    type="checkbox"
                    name="esAltaVital"
                    [(ngModel)]="esAltaVital"
                  />
                  <span>
                    <strong>Es una emergencia vital inmediata</strong> — Requiere atención prioritaria.
                  </span>
                </label>
              </div>
            </div>

            <div class="contingency-box">
              <label class="checkbox-row">
                <input
                  type="checkbox"
                  name="simularFalloMapas"
                  [(ngModel)]="simularFalloMapas"
                />
                <span>
                  <strong>Modo respaldo:</strong> buscar por zona o distrito si el mapa no está disponible.
                </span>
              </label>
            </div>

            <div class="form-submit-row">
              <button
                type="submit"
                class="btn btn-primary"
                [disabled]="loading() || cantidadReq < 1"
              >
                {{ loading() ? 'Buscando y reservando unidades...' : 'Buscar y reservar unidades' }}
              </button>
            </div>
          </form>
        </section>
      }
    </div>
  `,
  styles: [`
    .narrow-workspace { max-width: 780px; }
    .contingency-box {
      padding: 1rem 1.2rem;
      background-color: var(--bg-surface-muted);
      border: 1px solid var(--border-subtle);
      border-radius: var(--radius-input);
    }
    .form-submit-row { padding-top: 0.4rem; }
    .conflict-panel { display: flex; flex-direction: column; gap: 1.35rem; }
    .conflict-head { display: flex; flex-direction: column; gap: 0.35rem; }
    .conflict-id { font-size: 1.5rem; font-weight: 900; }
    .convocatoria-box {
      padding: 1.2rem 1.35rem;
      background-color: var(--brand-yellow-soft);
      border: 1px solid var(--brand-yellow);
      border-radius: var(--radius-subcard);
      display: flex;
      flex-direction: column;
      gap: 0.4rem;
    }
    .convocatoria-copy { font-size: 0.94rem; color: var(--text-primary); }
    .conflict-actions { display: flex; flex-wrap: wrap; gap: 0.75rem; }
  `],
})
export class EmergenciaNuevaComponent {
  private readonly emergenciasService = inject(EmergenciasService);
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);

  protected readonly bancos = BANCOS_PIURA_CATALOGO;
  protected readonly grupos = GRUPOS_SANGUINEOS;
  protected readonly factoresRh = FACTORES_RH;

  protected hospitalOrigenId = this.authService.currentUser()?.hospitalId || 'HOSP-MINSA-01';
  protected coordenadasOrigen = '-5.1945,-80.6328';
  protected tipoSangreReq: TipoSangre = 'O';
  protected factorRhReq: FactorRh = '+';
  protected cantidadReq = 2;
  protected esAltaVital = true;
  protected simularFalloMapas = false;

  protected readonly loading = signal(false);
  protected readonly apiError = signal<ExtractedError | null>(null);
  protected readonly conflictoSinStock = signal<EmergenciaSinStockConflictDTO | null>(null);

  onHospitalChange(bancoId: string): void {
    const found = this.bancos.find((b) => b.id === bancoId);
    if (found) {
      this.coordenadasOrigen = found.coordenadas;
    }
  }

  ejecutarMatching(): void {
    if (this.cantidadReq < 1) return;

    const payload: CrearEmergenciaRequest = {
      hospitalOrigenId: this.hospitalOrigenId,
      coordenadasOrigen: this.coordenadasOrigen.trim(),
      medicoId: this.authService.currentUser()?.id ?? 12,
      tipoSangreReq: this.tipoSangreReq,
      factorRhReq: this.factorRhReq,
      cantidadReq: Number(this.cantidadReq),
      urgencia: this.esAltaVital ? 'ALTA_VITAL' : 'MEDIA',
      simularFalloMapas: this.simularFalloMapas,
    };

    this.loading.set(true);
    this.apiError.set(null);
    this.conflictoSinStock.set(null);

    this.emergenciasService.crearSolicitudYMatching(payload).subscribe({
      next: (resp) => {
        this.loading.set(false);
        void this.router.navigate(['/emergencias', resp.solicitudId]);
      },
      error: (err: unknown) => {
        this.loading.set(false);
        if (err instanceof HttpErrorResponse && err.status === 409 && err.error?.solicitudId) {
          this.conflictoSinStock.set(err.error as EmergenciaSinStockConflictDTO);
          return;
        }
        this.apiError.set(extractApiError(err));
      },
    });
  }
}
