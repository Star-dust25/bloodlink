import { ChangeDetectionStrategy, Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import {
  BANCOS_PIURA_CATALOGO,
  FACTORES_RH,
  GRUPOS_SANGUINEOS,
} from '../../core/config/api.config';
import {
  EstadoUnidad,
  StockRedResponseDTO,
  UnidadSangreResponseDTO,
} from '../../core/models/inventario.model';
import { AuthService } from '../../core/services/auth.service';
import { InventarioService } from '../../core/services/inventario.service';
import { ExtractedError, extractApiError } from '../../core/utils/error.util';
import {
  formatAlertaCaducidad,
  formatNivelAbastecimiento,
} from '../../core/utils/label.util';
import { EmptyStateComponent } from '../../shared/ui/empty-state/empty-state.component';
import { InlineAlertComponent } from '../../shared/ui/inline-alert/inline-alert.component';
import { PageHeaderComponent } from '../../shared/ui/page-header/page-header.component';
import { StatusIndicatorComponent } from '../../shared/ui/status-indicator/status-indicator.component';

@Component({
  selector: 'app-inventario-list',
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
      eyebrow="INVENTARIO DE SANGRE"
      title="Stock en red e inventario"
      subtitle="Consulta las unidades disponibles en cada hospital y controla las fechas de vencimiento de cada bolsa."
    >
      @if (esPersonalBanco()) {
        <a routerLink="/inventario/nueva" class="btn btn-primary">
          Ingresar nueva unidad
        </a>
      }
    </app-page-header>

    <div class="tab-bar">
      <button
        type="button"
        class="tab-btn"
        [class.active]="activeTab() === 'STOCK_RED'"
        (click)="cambiarTab('STOCK_RED')"
      >
        Disponibilidad por hospital
      </button>
      <button
        type="button"
        class="tab-btn"
        [class.active]="activeTab() === 'UNIDADES'"
        (click)="cambiarTab('UNIDADES')"
      >
        Bolsas registradas ({{ unidades().length }})
      </button>
    </div>

    <div class="filter-bar">
      <div class="field">
        <label class="field-label" for="fBanco">Hospital</label>
        <select id="fBanco" class="select" [(ngModel)]="hospitalId" (ngModelChange)="recargar()">
          <option value="">Todos los hospitales</option>
          @for (b of bancos; track b.id) {
            <option [value]="b.id">{{ b.nombre }}</option>
          }
        </select>
      </div>

      <div class="field">
        <label class="field-label" for="fTipo">Grupo sanguíneo</label>
        <select id="fTipo" class="select" [(ngModel)]="tipoSangre" (ngModelChange)="recargar()">
          <option value="">Todos</option>
          @for (g of grupos; track g) {
            <option [value]="g">Grupo {{ g }}</option>
          }
        </select>
      </div>

      <div class="field">
        <label class="field-label" for="fRh">Factor Rh</label>
        <select id="fRh" class="select" [(ngModel)]="factorRh" (ngModelChange)="recargar()">
          <option value="">Ambos (+ / -)</option>
          @for (rh of factoresRh; track rh) {
            <option [value]="rh">Rh {{ rh }}</option>
          }
        </select>
      </div>

      @if (activeTab() === 'UNIDADES') {
        <div class="field">
          <label class="field-label" for="fEstado">Estado</label>
          <select id="fEstado" class="select" [(ngModel)]="estado" (ngModelChange)="recargar()">
            <option value="">Todos los estados</option>
            <option value="Disponible">Disponible</option>
            <option value="Reservada">Reservada</option>
            <option value="Transfundida">Transfundida</option>
            <option value="Caducada">Caducada</option>
          </select>
        </div>
      }
    </div>

    @if (apiError(); as err) {
      <div style="margin-bottom: 1.5rem;">
        <app-inline-alert variant="error" [title]="err.mensaje" [message]="err.detalle" />
      </div>
    }

    @if (mensajeExito(); as msg) {
      <div style="margin-bottom: 1.5rem;">
        <app-inline-alert variant="success" title="Inventario actualizado" [message]="msg" />
      </div>
    }

    <!-- PESTAÑA 1: STOCK EN RED -->
    @if (activeTab() === 'STOCK_RED' && stockRed(); as sr) {
      <div class="kpi-grid tabular">
        <div class="kpi-card">
          <span class="eyebrow">GRUPO CONSULTADO</span>
          <div class="kpi-value">{{ sr.tipoSangre }} / {{ sr.factorRh }}</div>
          <span class="kpi-sub">Actualizado al instante</span>
        </div>
        <div class="kpi-card">
          <span class="eyebrow">UNIDADES DISPONIBLES</span>
          <div class="kpi-value">{{ sr.stockGlobalDisponible }}</div>
          <span class="kpi-sub">Listas para usarse</span>
        </div>
        <div class="kpi-card">
          <span class="eyebrow">NIVEL DE ABASTECIMIENTO</span>
          <div style="margin-top: 0.45rem;">
            <app-status-indicator
              [tone]="sr.nivelAbastecimiento === 'ESTABLE' ? 'ok' : 'critical'"
              [label]="formatNivel(sr.nivelAbastecimiento)"
              [detail]="sr.nivelAbastecimiento === 'CRITICO' ? 'Conviene convocar donantes' : 'Capacidad adecuada'"
            />
          </div>
          <span class="kpi-sub">{{ sr.bancos.length }} hospitales en red</span>
        </div>
      </div>

      <div class="table-wrap">
        <table class="editorial-table tabular">
          <thead>
            <tr>
              <th>Establecimiento de salud</th>
              <th>Tipo de centro</th>
              <th>Zona / Distrito</th>
              <th>Disponibles</th>
              <th>Por vencer (&lt;= 7 días)</th>
            </tr>
          </thead>
          <tbody>
            @for (b of sr.bancos; track b.bancoId) {
              <tr>
                <td>
                  <div><strong>{{ b.nombre }}</strong></div>
                  <div class="code-mono" style="font-size: 0.78rem; color: var(--text-muted);">{{ b.bancoId }}</div>
                </td>
                <td>{{ b.tipoInstitucion }}</td>
                <td>{{ b.jurisdiccion }}</td>
                <td>
                  <strong style="font-size: 1.08rem;">{{ b.unidadesDisponibles }}</strong> u.
                </td>
                <td>
                  @if (b.proximasACaducar7Dias > 0) {
                    <app-status-indicator
                      tone="warning"
                      [label]="b.proximasACaducar7Dias + ' unidades'"
                      detail="Revisar pronto"
                    />
                  } @else {
                    <span style="color: var(--text-muted);">Ninguna</span>
                  }
                </td>
              </tr>
            }
          </tbody>
        </table>
      </div>
    }

    <!-- PESTAÑA 2: UNIDADES REGISTRADAS -->
    @if (activeTab() === 'UNIDADES') {
      @if (unidades().length === 0) {
        <app-empty-state
          eyebrow="INVENTARIO"
          title="No hay unidades que coincidan con tu búsqueda"
          description="Prueba cambiando el hospital, el grupo sanguíneo o el estado para ver otras bolsas."
        />
      } @else {
        <div class="table-wrap">
          <table class="editorial-table tabular">
            <thead>
              <tr>
                <th>Unidad / Código</th>
                <th>Banco</th>
                <th>Grupo</th>
                <th>Válida hasta</th>
                <th>Semáforo</th>
                <th>Estado</th>
                <th>Opciones</th>
              </tr>
            </thead>
            <tbody>
              @for (u of unidades(); track u.unidadId) {
                <tr>
                  <td>
                    <div class="code-mono"><strong>{{ u.unidadId }}</strong></div>
                    <div class="code-mono" style="font-size: 0.76rem; color: var(--text-muted);">
                      {{ u.codigoBarras }}
                    </div>
                  </td>
                  <td>{{ u.bancoId }}</td>
                  <td><strong>{{ u.tipoSangre }}{{ u.factorRh }}</strong> ({{ u.volumenMl }} mL)</td>
                  <td>{{ u.fechaCaducidad }}</td>
                  <td>
                    <app-status-indicator
                      [tone]="u.alertaCaducidad === 'ROJO' ? 'critical' : u.alertaCaducidad === 'AMARILLO' ? 'warning' : 'ok'"
                      [label]="formatAlerta(u.alertaCaducidad)"
                      [detail]="u.diasRestantes + ' d'"
                    />
                  </td>
                  <td><strong>{{ u.estado }}</strong></td>
                  <td>
                    <div class="row-actions">
                      @if (puedeVerTrazabilidad()) {
                        <a [routerLink]="['/trazabilidad', u.unidadId]" class="btn btn-ghost btn-sm">
                          Historial
                        </a>
                      }
                      @if (esPersonalBanco() && transicionesValidas(u.estado).length > 0) {
                        <select
                          class="select select-inline"
                          aria-label="Cambiar estado de unidad"
                          [ngModel]="''"
                          (ngModelChange)="cambiarEstadoUnidad(u, $event)"
                        >
                          <option value="" disabled>Cambiar estado...</option>
                          @for (dest of transicionesValidas(u.estado); track dest) {
                            <option [value]="dest">{{ dest }}</option>
                          }
                        </select>
                      }
                    </div>
                  </td>
                </tr>
              }
            </tbody>
          </table>
        </div>
      }
    }
  `,
  styles: [`
    .row-actions { display: flex; flex-wrap: wrap; align-items: center; gap: 0.5rem; }
    .select-inline {
      min-height: 36px; padding: 0.3rem 0.75rem; font-size: 0.82rem;
      border-radius: var(--radius-pill); width: auto;
    }
  `],
})
export class InventarioListComponent implements OnInit {
  private readonly inventarioService = inject(InventarioService);
  private readonly authService = inject(AuthService);

  protected readonly bancos = BANCOS_PIURA_CATALOGO;
  protected readonly grupos = GRUPOS_SANGUINEOS;
  protected readonly factoresRh = FACTORES_RH;

  protected readonly activeTab = signal<'STOCK_RED' | 'UNIDADES'>('STOCK_RED');
  protected readonly stockRed = signal<StockRedResponseDTO | null>(null);
  protected readonly unidades = signal<UnidadSangreResponseDTO[]>([]);
  protected readonly apiError = signal<ExtractedError | null>(null);
  protected readonly mensajeExito = signal<string>('');

  protected readonly formatAlerta = formatAlertaCaducidad;
  protected readonly formatNivel = formatNivelAbastecimiento;

  protected readonly esPersonalBanco = computed(
    () => this.authService.currentRole() === 'PERSONAL_BANCO'
  );
  protected readonly puedeVerTrazabilidad = computed(() =>
    this.authService.hasRole(['PERSONAL_BANCO', 'COORDINADOR'])
  );

  protected hospitalId = '';
  protected tipoSangre = '';
  protected factorRh = '';
  protected estado = '';

  ngOnInit(): void {
    this.recargar();
  }

  cambiarTab(tab: 'STOCK_RED' | 'UNIDADES'): void {
    this.activeTab.set(tab);
    this.mensajeExito.set('');
  }

  recargar(): void {
    this.apiError.set(null);
    const filtro = {
      hospitalId: this.hospitalId || undefined,
      tipoSangre: this.tipoSangre || undefined,
      factorRh: this.factorRh || undefined,
      estado: this.estado || undefined,
    };

    this.inventarioService.consultarStockRed(filtro).subscribe({
      next: (res) => this.stockRed.set(res),
      error: (err) => this.apiError.set(extractApiError(err)),
    });

    this.inventarioService.listarUnidades(filtro).subscribe({
      next: (list) => this.unidades.set(list),
      error: (err) => this.apiError.set(extractApiError(err)),
    });
  }

  transicionesValidas(estadoActual: EstadoUnidad): EstadoUnidad[] {
    switch (estadoActual) {
      case 'Disponible':
        return ['Reservada', 'Caducada'];
      case 'Reservada':
        return ['Transfundida', 'Disponible'];
      default:
        return [];
    }
  }

  cambiarEstadoUnidad(u: UnidadSangreResponseDTO, nuevoEstado: EstadoUnidad): void {
    if (!nuevoEstado) return;
    this.apiError.set(null);
    this.mensajeExito.set('');

    this.inventarioService
      .actualizarEstado(u.unidadId, {
        nuevoEstado,
        motivo: `Actualización desde el panel de inventario (${u.estado} -> ${nuevoEstado})`,
      })
      .subscribe({
        next: (actualizada) => {
          this.mensajeExito.set(
            `La unidad ${actualizada.unidadId} pasó al estado "${actualizada.estado}".`
          );
          this.recargar();
        },
        error: (err) => this.apiError.set(extractApiError(err)),
      });
  }
}
