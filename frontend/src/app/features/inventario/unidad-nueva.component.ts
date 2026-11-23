import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import {
  BANCOS_PIURA_CATALOGO,
  FACTORES_RH,
  GRUPOS_SANGUINEOS,
} from '../../core/config/api.config';
import {
  FactorRh,
  RegistrarUnidadRequest,
  TipoSangre,
  UnidadSangreResponseDTO,
} from '../../core/models/inventario.model';
import { AuthService } from '../../core/services/auth.service';
import { InventarioService } from '../../core/services/inventario.service';
import { ExtractedError, extractApiError } from '../../core/utils/error.util';
import { BarcodeCardComponent } from '../../shared/ui/barcode-card/barcode-card.component';
import { InlineAlertComponent } from '../../shared/ui/inline-alert/inline-alert.component';
import { PageHeaderComponent } from '../../shared/ui/page-header/page-header.component';

@Component({
  selector: 'app-unidad-nueva',
  standalone: true,
  imports: [
    FormsModule,
    RouterLink,
    PageHeaderComponent,
    InlineAlertComponent,
    BarcodeCardComponent,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="narrow-workspace">
      <app-page-header
        eyebrow="INGRESO AL INVENTARIO"
        title="Ingresar nueva unidad de sangre"
        subtitle="Registra una nueva bolsa donada. El sistema calcula automáticamente sus 42 días de vigencia y genera su etiqueta."
      >
        <a routerLink="/inventario" class="btn btn-ghost">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M15 18L9 12L15 6" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
          <span>Volver a Inventario</span>
        </a>
      </app-page-header>

      @if (apiError(); as err) {
        <div style="margin-bottom: 1.5rem;">
          <app-inline-alert
            variant="error"
            [title]="err.status === 422 ? 'No se puede ingresar una unidad vencida (más de 42 días desde su extracción)' : err.mensaje"
            [message]="err.detalle || err.mensaje"
          />
        </div>
      }

      @if (unidadCreada(); as u) {
        <app-barcode-card
          [unidad]="u"
          (registerAnother)="resetForm()"
        />
      } @else {
        <section class="surface-card">
          <form class="form-stack" (ngSubmit)="registrar()">
            <div class="form-row">
              <div class="field">
                <label class="field-label" for="bancoSelect">Banco que custodia la unidad</label>
                <select id="bancoSelect" name="bancoId" class="select" [(ngModel)]="bancoId">
                  @for (b of bancos; track b.id) {
                    <option [value]="b.id">{{ b.nombre }}</option>
                  }
                </select>
              </div>

              <div class="field">
                <label class="field-label" for="origenSelect">Tipo de donación</label>
                <select id="origenSelect" name="origen" class="select" [(ngModel)]="origen">
                  <option value="DONACION_VOLUNTARIA">Donación voluntaria</option>
                  <option value="CAMPANA_REGIONAL_MINSA">Campaña regional</option>
                  <option value="REPOSICION_FAMILIAR">Reposición familiar</option>
                </select>
              </div>
            </div>

            <div class="form-row">
              <div class="field">
                <span class="field-label">Grupo sanguíneo</span>
                <div class="selector-grid">
                  @for (g of grupos; track g) {
                    <button
                      type="button"
                      class="selector-btn tabular"
                      [class.selected]="tipoSangre === g"
                      (click)="tipoSangre = g"
                    >
                      {{ g }}
                    </button>
                  }
                </div>
              </div>

              <div class="field">
                <span class="field-label">Factor Rh</span>
                <div class="selector-grid">
                  @for (rh of factoresRh; track rh) {
                    <button
                      type="button"
                      class="selector-btn tabular"
                      [class.selected]="factorRh === rh"
                      (click)="factorRh = rh"
                    >
                      {{ rh }}
                    </button>
                  }
                </div>
              </div>
            </div>

            <div class="form-row">
              <div class="field">
                <label class="field-label" for="fechaExt">Fecha de extracción</label>
                <input
                  id="fechaExt"
                  name="fechaExtraccion"
                  type="date"
                  class="input tabular"
                  required
                  [(ngModel)]="fechaExtraccion"
                />
                <span class="field-hint">
                  Las unidades tienen una vigencia máxima de 42 días. Si ingresas una fecha vencida, el registro será bloqueado.
                </span>
              </div>

              <div class="field">
                <label class="field-label" for="volInput">Volumen (mL)</label>
                <input
                  id="volInput"
                  name="volumenMl"
                  type="number"
                  min="200"
                  max="600"
                  class="input tabular"
                  required
                  [(ngModel)]="volumenMl"
                />
              </div>
            </div>

            <div>
              <button
                type="submit"
                class="btn btn-primary"
                [disabled]="loading() || !fechaExtraccion"
              >
                {{ loading() ? 'Validando vigencia y generando etiqueta...' : 'Registrar unidad y generar etiqueta' }}
              </button>
            </div>
          </form>
        </section>
      }
    </div>
  `,
  styles: [`
    .narrow-workspace { max-width: 740px; }
  `],
})
export class UnidadNuevaComponent {
  private readonly inventarioService = inject(InventarioService);
  private readonly authService = inject(AuthService);

  protected readonly bancos = BANCOS_PIURA_CATALOGO;
  protected readonly grupos = GRUPOS_SANGUINEOS;
  protected readonly factoresRh = FACTORES_RH;

  protected bancoId = this.authService.currentUser()?.hospitalId || 'HOSP-MINSA-01';
  protected tipoSangre: TipoSangre = 'O';
  protected factorRh: FactorRh = '+';
  protected volumenMl = 450;
  protected fechaExtraccion = new Date().toISOString().slice(0, 10);
  protected origen = 'DONACION_VOLUNTARIA';

  protected readonly loading = signal(false);
  protected readonly apiError = signal<ExtractedError | null>(null);
  protected readonly unidadCreada = signal<UnidadSangreResponseDTO | null>(null);

  registrar(): void {
    if (!this.fechaExtraccion) return;

    const payload: RegistrarUnidadRequest = {
      bancoId: this.bancoId,
      tipoSangre: this.tipoSangre,
      factorRh: this.factorRh,
      volumenMl: Number(this.volumenMl) || 450,
      fechaExtraccion: this.fechaExtraccion,
      origen: this.origen,
    };

    this.loading.set(true);
    this.apiError.set(null);

    this.inventarioService.registrarUnidad(payload).subscribe({
      next: (creada) => {
        this.loading.set(false);
        this.unidadCreada.set(creada);
      },
      error: (err) => {
        this.loading.set(false);
        this.apiError.set(extractApiError(err));
      },
    });
  }

  resetForm(): void {
    this.unidadCreada.set(null);
    this.apiError.set(null);
    this.fechaExtraccion = new Date().toISOString().slice(0, 10);
  }
}
