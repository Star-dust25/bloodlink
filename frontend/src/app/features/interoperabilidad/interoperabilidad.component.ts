import { ChangeDetectionStrategy, Component, OnInit, inject, signal } from '@angular/core';
import { UDDIDirectorio } from '../../core/models/interoperabilidad.model';
import { InteroperabilidadService } from '../../core/services/interoperabilidad.service';
import { PageHeaderComponent } from '../../shared/ui/page-header/page-header.component';

@Component({
  selector: 'app-interoperabilidad',
  standalone: true,
  imports: [PageHeaderComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <app-page-header
      eyebrow="ARQUITECTURA SOA · SECCIÓN 11 DEL INFORME TÉCNICO"
      title="Interoperabilidad Híbrida SOAP 1.1 / WSDL / UDDI"
      subtitle="Especificación formal del bus ServicioBloodLinkSOAP para validación biométrica RENIEC y reserva transaccional en hospitales legados EsSalud."
    />

    <div class="tab-bar">
      <button
        type="button"
        class="tab-btn"
        [class.active]="activeTab() === 'UDDI'"
        (click)="activeTab.set('UDDI')"
      >
        Directorio UDDI (Páginas Blancas, Amarillas y Verdes)
      </button>
      <button
        type="button"
        class="tab-btn"
        [class.active]="activeTab() === 'WSDL'"
        (click)="activeTab.set('WSDL')"
      >
        Contrato WSDL 1.1 Oficial (XML)
      </button>
    </div>

    @if (activeTab() === 'UDDI') {
      @if (uddi(); as dir) {
        <div class="uddi-stack">
          <section class="surface-card">
            <span class="eyebrow">PÁGINAS BLANCAS · IDENTIDAD ORGANIZACIONAL</span>
            <h2 class="sec-title">{{ dir.paginasBlancas.organizacion }}</h2>
            <dl class="uddi-meta tabular">
              <div>
                <dt>Entidad</dt>
                <dd>{{ dir.paginasBlancas.entidad }}</dd>
              </div>
              <div>
                <dt>RUC Institucional</dt>
                <dd class="code-mono">{{ dir.paginasBlancas.ruc }}</dd>
              </div>
              <div>
                <dt>Jurisdicción</dt>
                <dd>{{ dir.paginasBlancas.jurisdiccion }}</dd>
              </div>
              <div>
                <dt>Namespace WSDL</dt>
                <dd class="code-mono">{{ dir.namespace }}</dd>
              </div>
            </dl>
          </section>

          <section class="surface-card">
            <span class="eyebrow">PÁGINAS AMARILLAS · CLASIFICACIÓN INDUSTRIAL UNSPSC</span>
            <div class="table-wrap" style="margin-top: 0.85rem;">
              <table class="editorial-table">
                <thead>
                  <tr>
                    <th>Categoría</th>
                    <th>Código UNSPSC</th>
                    <th>Estándar</th>
                  </tr>
                </thead>
                <tbody>
                  @for (am of dir.paginasAmarillas; track am.codigoUnspsc) {
                    <tr>
                      <td>{{ am.categoria }}</td>
                      <td class="code-mono">{{ am.codigoUnspsc }}</td>
                      <td class="code-mono">{{ am.estandar }}</td>
                    </tr>
                  }
                </tbody>
              </table>
            </div>
          </section>

          <section class="surface-card">
            <span class="eyebrow">PÁGINAS VERDES · BINDINGS TÉCNICOS Y OPERACIONES SOAP</span>
            <div class="table-wrap" style="margin-top: 0.85rem;">
              <table class="editorial-table">
                <thead>
                  <tr>
                    <th>Operación WSDL</th>
                    <th>Protocolo</th>
                    <th>SOAPAction</th>
                    <th>Descripción</th>
                  </tr>
                </thead>
                <tbody>
                  @for (v of dir.paginasVerdes; track v.operacion) {
                    <tr>
                      <td class="code-mono"><strong>{{ v.operacion }}</strong></td>
                      <td>{{ v.protocolo }}</td>
                      <td class="code-mono">{{ v.soapAction }}</td>
                      <td>{{ v.descripcion }}</td>
                    </tr>
                  }
                </tbody>
              </table>
            </div>
          </section>
        </div>
      } @else {
        <div class="surface-card">Cargando directorio UDDI desde <code class="code-mono">/api/v1/uddi/servicios</code>...</div>
      }
    }

    @if (activeTab() === 'WSDL') {
      <section class="ink-panel wsdl-panel">
        <div class="wsdl-toolbar">
          <div>
            <span class="eyebrow" style="color: rgba(246, 241, 234, 0.65);">GET /soap/v1/interoperabilidad?wsdl</span>
            <div class="code-mono">ServicioBloodLinkSOAP.wsdl</div>
          </div>
          <button type="button" class="btn btn-sm wsdl-copy-btn" (click)="copiarWsdl()">
            {{ copiado() ? 'WSDL Copiado ✓' : 'Copiar WSDL' }}
          </button>
        </div>
        <pre class="wsdl-code code-mono">{{ wsdlXml() || 'Cargando contrato WSDL...' }}</pre>
      </section>
    }
  `,
  styles: [`
    .uddi-stack {
      display: flex;
      flex-direction: column;
      gap: 1.5rem;
    }
    .sec-title {
      font-size: 1.2rem;
      font-weight: 600;
      margin: 0.35rem 0 1rem;
    }
    .uddi-meta {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
      gap: 1rem;
      padding-top: 1rem;
      border-top: 1px solid var(--border-subtle);
    }
    .uddi-meta dt {
      font-size: 0.72rem;
      text-transform: uppercase;
      letter-spacing: 0.06em;
      color: var(--text-muted);
    }
    .uddi-meta dd {
      font-size: 0.9rem;
      font-weight: 500;
      margin-top: 0.2rem;
    }
    .wsdl-panel {
      display: flex;
      flex-direction: column;
      gap: 1rem;
    }
    .wsdl-toolbar {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding-bottom: 0.85rem;
      border-bottom: 1px solid rgba(246, 241, 234, 0.18);
    }
    .wsdl-copy-btn {
      background-color: var(--text-on-ink);
      color: var(--bg-ink);
    }
    .wsdl-code {
      overflow-x: auto;
      font-size: 0.8rem;
      line-height: 1.55;
      color: var(--text-on-ink);
      max-height: 540px;
    }
  `],
})
export class InteroperabilidadComponent implements OnInit {
  private readonly interopService = inject(InteroperabilidadService);

  protected readonly activeTab = signal<'UDDI' | 'WSDL'>('UDDI');
  protected readonly uddi = signal<UDDIDirectorio | null>(null);
  protected readonly wsdlXml = signal<string>('');
  protected readonly copiado = signal(false);

  ngOnInit(): void {
    this.interopService.obtenerDirectorioUDDI().subscribe({
      next: (data) => this.uddi.set(data),
    });
    this.interopService.obtenerContratoWSDL().subscribe({
      next: (xml) => this.wsdlXml.set(xml),
    });
  }

  copiarWsdl(): void {
    const xml = this.wsdlXml();
    if (!xml || typeof navigator === 'undefined' || !navigator.clipboard) return;
    void navigator.clipboard.writeText(xml).then(() => {
      this.copiado.set(true);
      setTimeout(() => this.copiado.set(false), 2000);
    });
  }
}
