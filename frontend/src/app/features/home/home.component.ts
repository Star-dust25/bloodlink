import { AfterViewInit, ChangeDetectionStrategy, Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { HealthStatus } from '../../core/models/interoperabilidad.model';
import { AuthService } from '../../core/services/auth.service';
import { InteroperabilidadService } from '../../core/services/interoperabilidad.service';
import { StatusIndicatorComponent } from '../../shared/ui/status-indicator/status-indicator.component';

type GrupoCompleto = 'O-' | 'O+' | 'A-' | 'A+' | 'B-' | 'B+' | 'AB-' | 'AB+';

interface CompatibilidadInfo {
  donaA: readonly GrupoCompleto[];
  recibeDe: readonly GrupoCompleto[];
  nota: string;
}

const GRUPOS_COMPLETOS: readonly GrupoCompleto[] = [
  'O-',
  'O+',
  'A-',
  'A+',
  'B-',
  'B+',
  'AB-',
  'AB+',
];

const COMPATIBILIDAD_MAP: Record<GrupoCompleto, CompatibilidadInfo> = {
  'O-': {
    donaA: ['O-', 'O+', 'A-', 'A+', 'B-', 'B+', 'AB-', 'AB+'],
    recibeDe: ['O-'],
    nota: 'Donante universal: tu sangre ayuda en cualquier emergencia.',
  },
  'O+': {
    donaA: ['O+', 'A+', 'B+', 'AB+'],
    recibeDe: ['O-', 'O+'],
    nota: 'El grupo más frecuente en el Perú: fundamental en urgencias diarias.',
  },
  'A-': {
    donaA: ['A-', 'A+', 'AB-', 'AB+'],
    recibeDe: ['O-', 'A-'],
    nota: 'Muy valioso para pacientes A y AB con factor Rh negativo.',
  },
  'A+': {
    donaA: ['A+', 'AB+'],
    recibeDe: ['O-', 'O+', 'A-', 'A+'],
    nota: 'Segundo grupo más común en los hospitales de la región.',
  },
  'B-': {
    donaA: ['B-', 'B+', 'AB-', 'AB+'],
    recibeDe: ['O-', 'B-'],
    nota: 'Grupo poco frecuente: cada donación marca una gran diferencia.',
  },
  'B+': {
    donaA: ['B+', 'AB+'],
    recibeDe: ['O-', 'O+', 'B-', 'B+'],
    nota: 'Esencial para mantener reservas equilibradas en los bancos.',
  },
  'AB-': {
    donaA: ['AB-', 'AB+'],
    recibeDe: ['O-', 'A-', 'B-', 'AB-'],
    nota: 'Uno de los tipos más difíciles de ubicar ante una emergencia.',
  },
  'AB+': {
    donaA: ['AB+'],
    recibeDe: ['O-', 'O+', 'A-', 'A+', 'B-', 'B+', 'AB-', 'AB+'],
    nota: 'Receptor universal: puedes recibir sangre de cualquier grupo.',
  },
};

const HOSPITALES_RED = [
  {
    nombre: 'Hospital Nacional Santa Rosa',
    meta: 'MINSA · Piura Oeste',
    direccion: 'Av. Grau Km 3.5, Veintiséis de Octubre',
    acento: 'turquoise',
  },
  {
    nombre: 'Hospital III José Cayetano Heredia',
    meta: 'EsSalud · Castilla',
    direccion: 'Av. Independencia s/n, Urb. Miraflores',
    acento: 'yellow',
  },
  {
    nombre: 'Hospital de Apoyo II-2 Sullana',
    meta: 'MINSA · Sullana',
    direccion: 'Av. Santa Rosa s/n, Sullana',
    acento: 'turquoise',
  },
  {
    nombre: 'Clínica Miraflores',
    meta: 'Sector Privado · Castilla',
    direccion: 'Calle Las Dalias A-12, Urb. Miraflores',
    acento: 'yellow',
  },
] as const;

const MITOS_DONACION = [
  {
    mito: 'Donar sangre debilita o engorda.',
    verdad:
      'Tu cuerpo recupera el volumen de líquido en menos de 24 horas y no altera tu peso ni tu energía diaria.',
  },
  {
    mito: 'Puede haber riesgo de contagio.',
    verdad:
      'Todo el material utilizado es 100% nuevo, estéril y de un solo uso para cada persona.',
  },
  {
    mito: 'Toma demasiado tiempo del día.',
    verdad:
      'La extracción dura solo entre 10 y 15 minutos, y con tu registro previo en BloodLink el proceso es aún más ágil.',
  },
] as const;

const PREGUNTAS_FRECUENTES = [
  {
    pregunta: '¿Cuánto tiempo toma registrarme en BloodLink?',
    respuesta:
      'Menos de 3 minutos. Solo ingresas tu DNI, verificamos tus datos automáticamente y eliges tu tipo de sangre y medio de contacto.',
  },
  {
    pregunta: '¿Mis datos personales están seguros?',
    respuesta:
      'Sí. Tu DNI y teléfono se guardan protegidos bajo la Ley de Protección de Datos Personales y nunca se muestran completos a terceros.',
  },
  {
    pregunta: '¿Registrarme me obliga a donar de inmediato?',
    respuesta:
      'No. Quedas disponible como donante voluntario y solo te avisaremos si algún hospital cercano necesita tu tipo de sangre con urgencia.',
  },
  {
    pregunta: '¿Cómo ingresa el personal de un hospital?',
    respuesta:
      'A través del botón "Acceso hospitales" en la parte superior, utilizando las credenciales asignadas por su establecimiento.',
  },
] as const;

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [RouterLink, StatusIndicatorComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="home-flow">
      <!-- BLOQUE 1: HERO DE ALTO ENFOQUE -->
      <div class="hero-viewport">
        <section class="hero-card">
          <div class="hero-copy">
            <app-status-indicator
              [tone]="health()?.status === 'DEGRADED' ? 'warning' : 'ok'"
              label="Red de hospitales activa"
            />

            <h1 class="hero-title">Conectamos bancos de sangre para llegar a tiempo.</h1>
            <p class="hero-lead">
              Regístrate como donante voluntario en menos de 3 minutos o coordina unidades
              disponibles entre hospitales de Piura en tiempo real.
            </p>

            <div class="hero-actions">
              <a routerLink="/donantes/registro" class="btn btn-primary hero-cta">
                Quiero donar sangre
              </a>
              <a
                [routerLink]="auth.isAuthenticated() ? auth.defaultRouteForRole(auth.currentRole()) : '/login'"
                class="hero-text-link"
              >
                ¿Eres personal de salud? Ingresa aquí
              </a>
            </div>
          </div>

          <aside class="pillars-card" aria-label="Resumen de la red">
            <div class="pillars-head">
              <span>RED REGIONAL PIURA · TIEMPO REAL</span>
            </div>
            <div class="pillars-body">
              <div class="pillar-row">
                <strong class="pillar-title">Menos de 3 min</strong>
                <span class="pillar-desc">Registro voluntario solo con tu DNI.</span>
              </div>
              <div class="pillar-row">
                <strong class="pillar-title">4 centros en red</strong>
                <span class="pillar-desc">MINSA, EsSalud y clínicas de Piura conectados.</span>
              </div>
              <div class="pillar-row">
                <strong class="pillar-title">Atención 24/7</strong>
                <span class="pillar-desc">Búsqueda automática del banco más cercano ante emergencias.</span>
              </div>
            </div>
          </aside>
        </section>
      </div>

      <!-- BLOQUE 2: ¿CÓMO FUNCIONA? -->
      <section id="como-funciona" class="home-section">
        <div class="section-head">
          <span class="eyebrow">CÓMO FUNCIONA</span>
          <h2 class="section-title">Simple para donar, rápido en emergencias</h2>
          <p class="section-sub">Cada minuto cuenta cuando un paciente necesita una transfusión.</p>
        </div>

        <div class="three-grid">
          <article class="step-card bar-yellow">
            <span class="step-num c-yellow tabular">01</span>
            <h3 class="card-h3">Registro en minutos</h3>
            <p class="card-p">
              Los donantes validan su DNI y los bancos mantienen su stock actualizado día a día.
            </p>
          </article>

          <article class="step-card bar-turquoise">
            <span class="step-num c-turquoise tabular">02</span>
            <h3 class="card-h3">Ubicación inmediata</h3>
            <p class="card-p">
              Ante una emergencia, el sistema ubica al instante el hospital cercano con sangre compatible.
            </p>
          </article>

          <article class="step-card bar-sky">
            <span class="step-num c-sky tabular">03</span>
            <h3 class="card-h3">Reserva y aviso</h3>
            <p class="card-p">
              Se separa la unidad al momento y, si falta stock, convocamos automáticamente a donantes compatibles.
            </p>
          </article>
        </div>
      </section>

      <!-- BLOQUE 3: GUÍA INTERACTIVA DE COMPATIBILIDAD -->
      <section id="compatibilidad" class="home-section compat-card">
        <div class="section-head">
          <span class="eyebrow">GUÍA INTERACTIVA</span>
          <h2 class="section-title">Conoce el impacto de tu tipo de sangre</h2>
          <p class="section-sub">
            Selecciona tu grupo sanguíneo para ver a quiénes puedes donar y de quiénes puedes recibir.
          </p>
        </div>

        <div class="group-pills" role="group" aria-label="Seleccionar grupo sanguíneo">
          @for (g of grupos; track g) {
            <button
              type="button"
              class="group-pill tabular"
              [class.active]="grupoSeleccionado() === g"
              [attr.aria-pressed]="grupoSeleccionado() === g"
              (click)="grupoSeleccionado.set(g)"
            >
              {{ g }}
            </button>
          }
        </div>

        <div class="compat-panels">
          <div class="compat-box box-turquoise">
            @for (sel of [grupoSeleccionado()]; track sel) {
              <div class="compat-animated">
                <span class="eyebrow">PUEDES DONAR A ({{ sel }})</span>
                <div class="compat-groups tabular">
                  @for (target of compatibilidadActual().donaA; track target) {
                    <span class="compat-chip">{{ target }}</span>
                  }
                </div>
                <p class="compat-note">{{ compatibilidadActual().nota }}</p>
              </div>
            }
          </div>

          <div class="compat-box box-yellow">
            @for (sel of [grupoSeleccionado()]; track sel) {
              <div class="compat-animated">
                <span class="eyebrow">PUEDES RECIBIR DE ({{ sel }})</span>
                <div class="compat-groups tabular">
                  @for (src of compatibilidadActual().recibeDe; track src) {
                    <span class="compat-chip">{{ src }}</span>
                  }
                </div>
                <p class="compat-note">
                  Compatibilidad médica según grupo ABO y factor Rh para transfusiones seguras.
                </p>
              </div>
            }
          </div>
        </div>
      </section>

      <!-- BLOQUE 4: RED DE HOSPITALES CONECTADOS EN PIURA -->
      <section id="hospitales" class="home-section">
        <div class="section-head">
          <span class="eyebrow">RED REGIONAL</span>
          <h2 class="section-title">Bancos de sangre conectados en la red</h2>
          <p class="section-sub">
            Trabajamos en conjunto con establecimientos públicos y privados para que ninguna
            emergencia pierda minutos valiosos.
          </p>
        </div>

        <div class="hospitals-grid">
          @for (h of hospitales; track h.nombre) {
            <article class="hospital-card" [class.accent-yellow]="h.acento === 'yellow'">
              <div class="hospital-meta">{{ h.meta }}</div>
              <h3 class="card-h3">{{ h.nombre }}</h3>
              <p class="hospital-addr">{{ h.direccion }}</p>
            </article>
          }
        </div>
      </section>

      <!-- BLOQUE 5: MITOS Y VERDADES SOBRE DONAR SANGRE -->
      <section id="mitos" class="home-section">
        <div class="section-head">
          <span class="eyebrow">INFORMACIÓN CLARA</span>
          <h2 class="section-title">Despeja tus dudas antes de donar</h2>
          <p class="section-sub">
            Conoce la realidad detrás de las creencias más comunes sobre la donación voluntaria.
          </p>
        </div>

        <div class="three-grid">
          @for (m of mitos; track m.mito) {
            <article class="myth-card">
              <div class="myth-box">
                <span class="myth-label">Mito común</span>
                <span>{{ m.mito }}</span>
              </div>
              <div class="truth-box">
                <span class="truth-label">La verdad</span>
                <span>{{ m.verdad }}</span>
              </div>
            </article>
          }
        </div>
      </section>

      <!-- BLOQUE 6: REQUISITOS Y ACORDEÓN FLUIDO DE PREGUNTAS FRECUENTES -->
      <section id="requisitos" class="home-section">
        <div class="two-col">
          <div class="surface-card">
            <div class="section-head" style="margin-bottom: 1.25rem;">
              <span class="eyebrow">ANTES DE ACUDIR</span>
              <h2 class="req-title">Requisitos básicos para donar</h2>
            </div>
            <ul class="req-list">
              <li>Tener entre 18 y 60 años y contar con tu DNI vigente.</li>
              <li>Pesar más de 50 kg y gozar de buena salud general.</li>
              <li>Haber descansado bien y desayunado ligero antes de acudir al banco.</li>
              <li>No haber donado sangre en los últimos 3 meses (hombres) o 4 meses (mujeres).</li>
            </ul>
          </div>

          <div class="faq-col">
            <div class="section-head" style="margin-bottom: 0.5rem;">
              <span class="eyebrow">DUDAS COMUNES</span>
              <h2 class="req-title">Preguntas frecuentes</h2>
            </div>

            @for (item of faqs; track item.pregunta; let i = $index) {
              <div class="faq-item" [class.open]="faqAbierta() === i">
                <button
                  type="button"
                  class="faq-trigger"
                  [attr.aria-expanded]="faqAbierta() === i"
                  (click)="toggleFaq(i)"
                >
                  <span>{{ item.pregunta }}</span>
                  <span class="faq-chevron" aria-hidden="true">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                      <path
                        d="M6 9L12 15L18 9"
                        stroke="currentColor"
                        stroke-width="2.6"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      />
                    </svg>
                  </span>
                </button>
                <div class="faq-body-wrap">
                  <div class="faq-body-inner">
                    <p class="faq-a">{{ item.respuesta }}</p>
                  </div>
                </div>
              </div>
            }
          </div>
        </div>
      </section>

      <!-- BLOQUE 7: CIERRE FINAL ENFOCADO -->
      <section class="closing-banner">
        <div class="closing-copy">
          <h2 class="closing-title">Tu decisión hoy puede salvar hasta tres vidas.</h2>
          <p class="closing-sub">
            Únete gratis a la red de donantes voluntarios de Piura solo con tu DNI.
          </p>
        </div>
        <a routerLink="/donantes/registro" class="btn btn-primary hero-cta">
          Registrarme ahora
        </a>
      </section>
    </div>
  `,
  styles: [`
    .home-flow { display: flex; flex-direction: column; gap: 4rem; }
    .hero-card {
      width: 100%;
      background: linear-gradient(135deg, var(--brand-turquoise-soft) 0%, #FFFFFF 52%, var(--brand-yellow-soft) 100%);
      border: 1px solid var(--border-subtle); border-radius: 32px;
      padding: 2.75rem; box-shadow: var(--shadow-soft);
      display: grid; grid-template-columns: 1.25fr 0.9fr; align-items: center; gap: 2.5rem;
    }
    .hero-copy { display: flex; flex-direction: column; gap: 1.2rem; }
    .hero-title { font-size: clamp(2rem, 3.6vw, 2.95rem); font-weight: 900; }
    .hero-lead { font-size: 1.05rem; color: var(--text-secondary); max-width: 580px; }
    .hero-actions { display: flex; flex-wrap: wrap; align-items: center; gap: 1.25rem; margin-top: 0.4rem; }
    .hero-cta { padding: 0.9rem 1.95rem; min-height: 54px; font-size: 1.02rem; }
    .hero-text-link { font-size: 0.96rem; font-weight: 800; color: var(--brand-turquoise-dark); }
    .hero-text-link:hover { text-decoration: underline; text-underline-offset: 4px; opacity: 1; }
    .pillars-card {
      background-color: var(--bg-surface); border: 1px solid var(--border-subtle);
      border-radius: 28px; box-shadow: var(--shadow-soft); overflow: hidden;
    }
    .pillars-head {
      background-color: var(--brand-turquoise-dark); color: #FFFFFF;
      padding: 0.95rem 1.4rem; font-size: 0.76rem; font-weight: 800; letter-spacing: 0.08em;
      border-bottom: 4px solid var(--brand-yellow);
    }
    .pillars-body { padding: 0.5rem 1.4rem; }
    .pillar-row {
      display: flex; flex-direction: column; gap: 0.15rem; padding: 1rem 0;
      border-bottom: 1px solid var(--border-subtle);
    }
    .pillar-row:last-child { border-bottom: none; }
    .pillar-title { font-size: 1.06rem; font-weight: 900; color: var(--text-primary); }
    .pillar-desc { font-size: 0.92rem; color: var(--text-secondary); }
    .three-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1.35rem; }
    .step-card {
      background-color: var(--bg-surface); border: 1px solid var(--border-subtle);
      border-radius: 28px; padding: 1.85rem; box-shadow: var(--shadow-soft);
      display: flex; flex-direction: column; gap: 0.65rem;
    }
    .bar-yellow { border-top: 5px solid var(--brand-yellow); }
    .bar-turquoise { border-top: 5px solid var(--brand-turquoise); }
    .bar-sky { border-top: 5px solid var(--brand-sky); }
    .step-num { font-size: 2rem; font-weight: 900; line-height: 1; }
    .c-yellow { color: #C5910B; }
    .c-turquoise { color: var(--brand-turquoise-dark); }
    .c-sky { color: var(--brand-sky); }
    .card-h3 { font-size: 1.15rem; font-weight: 800; }
    .card-p { font-size: 0.96rem; color: var(--text-secondary); }
    @media (max-width: 860px) {
      .hero-card { grid-template-columns: 1fr; padding: 1.75rem; }
    }
  `],
})
export class HomeComponent implements OnInit, AfterViewInit {
  private readonly interopService = inject(InteroperabilidadService);
  protected readonly auth = inject(AuthService);

  protected readonly health = signal<HealthStatus | null>(null);
  protected readonly grupos = GRUPOS_COMPLETOS;
  protected readonly hospitales = HOSPITALES_RED;
  protected readonly mitos = MITOS_DONACION;
  protected readonly faqs = PREGUNTAS_FRECUENTES;
  protected readonly grupoSeleccionado = signal<GrupoCompleto>('O-');
  protected readonly faqAbierta = signal<number | null>(null);

  protected readonly compatibilidadActual = computed(
    () => COMPATIBILIDAD_MAP[this.grupoSeleccionado()]
  );

  ngOnInit(): void {
    this.interopService.verificarHealth().subscribe({
      next: (status) => this.health.set(status),
      error: () => this.health.set(null),
    });
  }

  ngAfterViewInit(): void {
    if (typeof window === 'undefined') return;
    const state = window.history.state as { scrollTo?: string } | null;
    const targetId = state?.scrollTo;
    if (targetId) {
      setTimeout(() => {
        document.getElementById(targetId)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
        window.history.replaceState({}, '', '/');
      }, 50);
    } else if (window.location.hash) {
      window.history.replaceState({}, '', '/');
    }
  }

  toggleFaq(index: number): void {
    this.faqAbierta.update((curr) => (curr === index ? null : index));
  }
}
