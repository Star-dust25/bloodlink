import { ChangeDetectionStrategy, Component, OnInit, inject, input, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TrazabilidadUnidadReporte } from '../../core/models/reporte.model';
import { ReportesService } from '../../core/services/reportes.service';
import { triggerBlobDownload } from '../../core/utils/download.util';
import { ExtractedError, extractApiError } from '../../core/utils/error.util';
import { formatAccionAuditoria, formatOrigenUnidad } from '../../core/utils/label.util';
import { EmptyStateComponent } from '../../shared/ui/empty-state/empty-state.component';
import { InlineAlertComponent } from '../../shared/ui/inline-alert/inline-alert.component';
import { PageHeaderComponent } from '../../shared/ui/page-header/page-header.component';

const UNIDADES_EJEMPLO = ['UND-2026-0401', 'UND-2026-0106', 'UND-2026-0101', 'UND-2026-0201'] as const;

@Component({
  selector: 'app-trazabilidad',
  standalone: true,
  imports: [FormsModule, PageHeaderComponent, InlineAlertComponent, EmptyStateComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <app-page-header
      eyebrow="HISTORIAL CLÍNICO"
      title="Historial de unidad"
      subtitle="Consulta el recorrido completo de una bolsa de sangre desde su donación hasta su uso, y descarga su certificado en PDF o Excel (CSV)."
    />

    <form class="filter-bar" (ngSubmit)="buscar()">
      <div class="field">
        <label class="field-label" for="unidadSearch">Código de unidad</label>
        <input
          id="unidadSearch"
          name="queryUnidadId"
          type="text"
          class="input code-mono tabular"
          placeholder="Ej. UND-2026-0401"
          required
          [(ngModel)]="queryUnidadId"
        />
      </div>
      <div>
        <button type="submit" class="btn btn-primary" [disabled]="loading() || !queryUnidadId.trim()">
          {{ loading() ? 'Buscando...' : 'Consultar historial' }}
        </button>
      </div>
    </form>

    <div class="presets-bar tabular">
      <span class="eyebrow">EJEMPLOS RÁPIDOS:</span>
      @for (ej of ejemplos; track ej) {
        <button type="button" class="btn btn-ghost btn-sm code-mono" (click)="seleccionarEjemplo(ej)">
          {{ ej }}
        </button>
      }
    </div>

    @if (apiError(); as err) {
      <div style="margin-bottom: 1.5rem;">
        <app-inline-alert variant="error" [title]="err.mensaje" [message]="err.detalle" />
      </div>
    }

    @if (reporte(); as rep) {
      <section class="surface-card" style="margin-bottom: 1.5rem;">
        <div class="rep-header">
          <div>
            <span class="eyebrow">DETALLE DE LA UNIDAD</span>
            <h2 class="rep-id tabular">{{ rep.unidadId }} · {{ rep.tipoSangre }}{{ rep.factorRh }}</h2>
            <p class="rep-sub code-mono">Código de barras: {{ rep.codigoBarras }}</p>
          </div>
          <div class="export-actions no-print">
            <button type="button" class="btn btn-primary btn-sm" (click)="exportar('pdf')">
              Descargar certificado PDF
            </button>
            <button type="button" class="btn btn-secondary btn-sm" (click)="exportar('csv')">
              Descargar para Excel (CSV)
            </button>
          </div>
        </div>

        <dl class="rep-meta tabular">
          <div>
            <dt>Banco custodio</dt>
            <dd>{{ rep.bancoNombre }}</dd>
          </div>
          <div>
            <dt>Estado actual</dt>
            <dd><strong>{{ rep.estadoActual }}</strong></dd>
          </div>
          <div>
            <dt>Fecha de extracción</dt>
            <dd>{{ rep.fechaExtraccion }}</dd>
          </div>
          <div>
            <dt>Válida hasta (42 días)</dt>
            <dd>{{ rep.fechaCaducidad }}</dd>
          </div>
          <div>
            <dt>Volumen y origen</dt>
            <dd>{{ rep.volumenMl }} mL · {{ formatOrigen(rep.origen) }}</dd>
          </div>
          <div>
            <dt>Movimientos registrados</dt>
            <dd>{{ rep.historial.length }} registros</dd>
          </div>
        </dl>
      </section>

      <section class="surface-card">
        <span class="eyebrow">RECORRIDO COMPLETO DE LA BOLSA</span>
        <div class="timeline tabular">
          @for (ev of rep.historial; track ev.id) {
            <div class="timeline-item">
              <div class="timeline-dot"></div>
              <div class="timeline-body">
                <div class="timeline-top">
                  <strong>{{ formatAccion(ev.accion) }}</strong>
                  <span class="timeline-time code-mono">{{ ev.timestamp }}</span>
                </div>
                <div class="timeline-transition">
                  De {{ ev.estadoAnterior }} a <strong>{{ ev.estadoNuevo }}</strong>
                </div>
                <p class="timeline-detail">{{ ev.detalle }}</p>
              </div>
            </div>
          }
        </div>
      </section>
    } @else if (!loading()) {
      <app-empty-state
        eyebrow="HISTORIAL"
        title="Ingresa el código de una unidad"
        description="Podrás revisar cada paso de su recorrido y descargar el certificado oficial en PDF o Excel."
      />
    }
  `,
  styles: [`
    .presets-bar { display: flex; flex-wrap: wrap; align-items: center; gap: 0.5rem; margin-bottom: 1.5rem; }
    .rep-header {
      display: flex; flex-wrap: wrap; align-items: flex-start; justify-content: space-between;
      gap: 1rem; padding-bottom: 1.25rem; border-bottom: 1px solid var(--border-subtle);
    }
    .rep-id { font-size: 1.5rem; font-weight: 900; margin-top: 0.2rem; }
    .rep-sub { font-size: 0.85rem; color: var(--text-muted); }
    .export-actions { display: flex; flex-wrap: wrap; gap: 0.6rem; }
    .rep-meta { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; padding-top: 1.15rem; }
    .rep-meta dt { font-size: 0.78rem; font-weight: 700; color: var(--text-muted); }
    .rep-meta dd { font-size: 0.94rem; font-weight: 700; margin-top: 0.2rem; }
    .timeline {
      margin-top: 1.25rem; display: flex; flex-direction: column; gap: 1.25rem;
      border-left: 2px solid var(--brand-turquoise); margin-left: 0.5rem; padding-left: 1.35rem;
    }
    .timeline-item { position: relative; }
    .timeline-dot {
      position: absolute; left: -1.72rem; top: 0.35rem; width: 11px; height: 11px;
      border-radius: 50%; background-color: var(--brand-yellow); border: 2px solid var(--brand-turquoise-dark);
    }
    .timeline-top { display: flex; flex-wrap: wrap; justify-content: space-between; gap: 0.5rem; font-size: 0.96rem; }
    .timeline-time { font-size: 0.78rem; color: var(--text-muted); }
    .timeline-transition { font-size: 0.84rem; color: var(--brand-turquoise-dark); font-weight: 700; margin: 0.2rem 0; }
    .timeline-detail { font-size: 0.9rem; color: var(--text-secondary); }
  `],
})
export class TrazabilidadComponent implements OnInit {
  private readonly reportesService = inject(ReportesService);

  readonly unidadId = input<string>('');

  protected readonly ejemplos = UNIDADES_EJEMPLO;
  protected readonly formatOrigen = formatOrigenUnidad;
  protected readonly formatAccion = formatAccionAuditoria;
  protected queryUnidadId = 'UND-2026-0401';
  protected readonly reporte = signal<TrazabilidadUnidadReporte | null>(null);
  protected readonly loading = signal(false);
  protected readonly apiError = signal<ExtractedError | null>(null);

  ngOnInit(): void {
    const paramId = this.unidadId();
    if (paramId) {
      this.queryUnidadId = paramId;
    }
    this.buscar();
  }

  seleccionarEjemplo(id: string): void {
    this.queryUnidadId = id;
    this.buscar();
  }

  buscar(): void {
    const id = this.queryUnidadId.trim();
    if (!id) return;

    this.loading.set(true);
    this.apiError.set(null);

    this.reportesService.obtenerTrazabilidadJSON(id).subscribe({
      next: (data) => {
        this.loading.set(false);
        this.reporte.set(data);
      },
      error: (err) => {
        this.loading.set(false);
        this.reporte.set(null);
        this.apiError.set(extractApiError(err));
      },
    });
  }

  exportar(format: 'pdf' | 'csv'): void {
    const rep = this.reporte();
    if (!rep) return;

    this.reportesService.descargarTrazabilidadArchivo(rep.unidadId, format).subscribe({
      next: (blob) => {
        triggerBlobDownload(blob, `trazabilidad_${rep.unidadId}.${format}`);
      },
      error: (err) => this.apiError.set(extractApiError(err)),
    });
  }
}
