import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';
import { UnidadSangreResponseDTO } from '../../../core/models/inventario.model';
import { formatAlertaCaducidad, formatOrigenUnidad } from '../../../core/utils/label.util';
import { StatusIndicatorComponent } from '../status-indicator/status-indicator.component';

@Component({
  selector: 'app-barcode-card',
  standalone: true,
  imports: [StatusIndicatorComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <section class="barcode-sheet surface-card-strong">
      <div class="sheet-top">
        <div>
          <span class="eyebrow">ETIQUETA DE UNIDAD REGISTRADA</span>
          <h2 class="sheet-id tabular">{{ unidad().unidadId }}</h2>
        </div>
        <div class="blood-group-stamp tabular">
          {{ unidad().tipoSangre }}{{ unidad().factorRh }}
        </div>
      </div>

      <div class="barcode-bars" aria-hidden="true">
        <span></span><span></span><span></span><span></span><span></span>
        <span></span><span></span><span></span><span></span><span></span>
        <span></span><span></span><span></span><span></span><span></span>
        <span></span><span></span><span></span><span></span><span></span>
      </div>
      <div class="barcode-code code-mono tabular">{{ unidad().codigoBarras }}</div>

      <dl class="meta-grid tabular">
        <div>
          <dt>Banco de sangre</dt>
          <dd>{{ unidad().bancoId }}</dd>
        </div>
        <div>
          <dt>Volumen</dt>
          <dd>{{ unidad().volumenMl }} mL</dd>
        </div>
        <div>
          <dt>Fecha de extracción</dt>
          <dd>{{ unidad().fechaExtraccion }}</dd>
        </div>
        <div>
          <dt>Válida hasta (42 días)</dt>
          <dd>{{ unidad().fechaCaducidad }}</dd>
        </div>
        <div>
          <dt>Origen</dt>
          <dd>{{ formatOrigen(unidad().origen) }}</dd>
        </div>
        <div>
          <dt>Vigencia</dt>
          <dd>
            <app-status-indicator
              [tone]="unidad().alertaCaducidad === 'ROJO' ? 'critical' : unidad().alertaCaducidad === 'AMARILLO' ? 'warning' : 'ok'"
              [label]="formatAlerta(unidad().alertaCaducidad)"
              [detail]="unidad().diasRestantes + ' días'"
            />
          </dd>
        </div>
      </dl>

      <div class="sheet-actions no-print">
        <button type="button" class="btn btn-primary" (click)="printLabel()">
          Imprimir etiqueta
        </button>
        <button type="button" class="btn btn-secondary" (click)="registerAnother.emit()">
          Registrar otra unidad
        </button>
      </div>
    </section>
  `,
  styles: [`
    .barcode-sheet {
      max-width: 580px;
      display: flex;
      flex-direction: column;
      gap: 1.35rem;
    }
    .sheet-top {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      gap: 1rem;
      padding-bottom: 1rem;
      border-bottom: 1px solid var(--border-subtle);
    }
    .sheet-id {
      font-size: 1.4rem;
      font-weight: 900;
      margin-top: 0.25rem;
    }
    .blood-group-stamp {
      font-size: 1.9rem;
      font-weight: 900;
      padding: 0.45rem 1.05rem;
      background-color: var(--brand-yellow);
      color: var(--brand-ink);
      border-radius: var(--radius-subcard);
      line-height: 1.1;
    }
    .barcode-bars {
      display: flex;
      align-items: stretch;
      height: 54px;
      gap: 4px;
      padding: 0.55rem 1rem;
      background-color: var(--bg-surface-muted);
      border-radius: var(--radius-input);
    }
    .barcode-bars span { flex: 1; background-color: var(--brand-ink); border-radius: 2px; }
    .barcode-bars span:nth-child(even) { flex: 2; }
    .barcode-bars span:nth-child(3n) { opacity: 0.28; }
    .barcode-code {
      text-align: center;
      font-size: 0.92rem;
      font-weight: 700;
      letter-spacing: 0.12em;
    }
    .meta-grid {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 0.9rem 1.25rem;
      padding-top: 0.85rem;
      border-top: 1px solid var(--border-subtle);
    }
    .meta-grid dt {
      font-size: 0.76rem;
      font-weight: 700;
      color: var(--text-muted);
    }
    .meta-grid dd {
      font-size: 0.94rem;
      font-weight: 700;
      color: var(--text-primary);
    }
    .sheet-actions {
      display: flex;
      flex-wrap: wrap;
      gap: 0.75rem;
      padding-top: 0.5rem;
    }
  `],
})
export class BarcodeCardComponent {
  readonly unidad = input.required<UnidadSangreResponseDTO>();
  readonly registerAnother = output<void>();

  protected readonly formatOrigen = formatOrigenUnidad;
  protected readonly formatAlerta = formatAlertaCaducidad;

  printLabel(): void {
    if (typeof window !== 'undefined') {
      window.print();
    }
  }
}
