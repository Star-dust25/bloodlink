import { ChangeDetectionStrategy, Component, DestroyRef, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { FACTORES_RH, GRUPOS_SANGUINEOS } from '../../core/config/api.config';
import {
  CiudadanoReniec,
  DonanteResponseDTO,
  RegistrarDonanteRequest,
} from '../../core/models/donante.model';
import { FactorRh, TipoSangre } from '../../core/models/inventario.model';
import { DonantesService } from '../../core/services/donantes.service';
import { ExtractedError, extractApiError } from '../../core/utils/error.util';
import { ConfirmModalComponent } from '../../shared/ui/confirm-modal/confirm-modal.component';

type PasoRegistro = 'IDENTIDAD' | 'DATOS_CONTACTO' | 'CONFIRMADO';

interface HospitalOpcion {
  id: string;
  nombre: string;
  detalle: string;
}

const DNIS_PRUEBA = [
  { dni: '76543210', nota: 'Mayor de edad' },
  { dni: '45678901', nota: 'Mayor de edad' },
  { dni: '61112222', nota: 'Menor de edad' },
] as const;

const HOSPITALES_OPCIONES: readonly HospitalOpcion[] = [
  {
    id: 'HOSP-MINSA-01',
    nombre: 'Hospital Santa Rosa',
    detalle: 'MINSA · Piura Oeste',
  },
  {
    id: 'HOSP-ESSALUD-04',
    nombre: 'Hospital Cayetano Heredia',
    detalle: 'EsSalud · Castilla',
  },
  {
    id: 'HOSP-MINSA-02',
    nombre: 'Hospital de Apoyo Sullana',
    detalle: 'MINSA · Sullana',
  },
  {
    id: 'CLIN-PRIV-01',
    nombre: 'Clínica Miraflores',
    detalle: 'Privada · Castilla',
  },
] as const;

const CELULAR_REGEX = /^9\d{8}$/;
const CORREO_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;

@Component({
  selector: 'app-registro-donante',
  standalone: true,
  imports: [FormsModule, RouterLink, ConfirmModalComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
  host: {
    '(document:click)': 'cerrarTooltips()',
    '(document:keydown.escape)': 'cerrarTooltips()',
  },
  template: `
    <div class="donor-page">
      <header class="donor-topbar">
        <a href="/" class="donor-brand" (click)="intentarSalir($event)">
          <span class="brand-dot" aria-hidden="true"></span>
          <span>BloodLink</span>
        </a>
        <a href="/" class="back-link" (click)="intentarSalir($event)">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M15 18L9 12L15 6" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
          <span>Volver</span>
        </a>
      </header>

      <main class="donor-card">
        @if (paso() !== 'CONFIRMADO') {
          <div class="progress-row tabular">
            <div class="progress-track" aria-hidden="true">
              <div
                class="progress-fill"
                [style.width.%]="paso() === 'IDENTIDAD' ? 50 : 100"
              ></div>
            </div>
            <span class="progress-label">
              {{ paso() === 'IDENTIDAD' ? 'Paso 1 de 2' : 'Paso 2 de 2' }}
            </span>
          </div>
        }

        <!-- PASO 1: VERIFICACIÓN DE IDENTIDAD -->
        @if (paso() === 'IDENTIDAD') {
          <section class="step-fade">
            <h1 class="step-title">Verifica tu identidad</h1>
            <p class="step-sub">Ingresa tu DNI para completar tus datos automáticamente.</p>

            <form class="form-stack" (ngSubmit)="validarIdentidad()">
              <div class="field">
                <div class="label-with-tip">
                  <label class="field-label" for="dniInput">Número de DNI</label>

                  <div class="tip-anchor" [class.open]="tooltipAbierto()" (click)="$event.stopPropagation()">
                    <button
                      type="button"
                      class="tip-btn"
                      aria-label="Ver DNIs de prueba"
                      [attr.aria-expanded]="tooltipAbierto()"
                      (click)="toggleTooltipDni($event)"
                    >
                      ?
                    </button>

                    <div class="tip-popover" role="tooltip">
                      <div class="tip-head">DNIs de prueba:</div>
                      @for (item of dnisPrueba; track item.dni) {
                        <button
                          type="button"
                          class="tip-option tabular"
                          (click)="seleccionarDniPrueba(item.dni)"
                        >
                          <strong>{{ item.dni }}</strong> · {{ item.nota }}
                        </button>
                      }
                    </div>
                  </div>
                </div>

                <input
                  id="dniInput"
                  name="dni"
                  type="text"
                  inputmode="numeric"
                  maxlength="8"
                  pattern="[0-9]{8}"
                  class="input dni-input tabular"
                  [class.input-error]="!!apiError()"
                  placeholder="Ej. 76543210"
                  required
                  [(ngModel)]="dni"
                  (ngModelChange)="apiError.set(null)"
                />
                @if (apiError(); as err) {
                  <p class="field-error" role="alert">{{ formatErrorText(err) }}</p>
                }
              </div>

              <button
                type="submit"
                class="btn btn-primary btn-full"
                [disabled]="loading() || dni.trim().length !== 8"
              >
                {{ loading() ? 'Validando...' : 'Validar DNI' }}
              </button>
            </form>
          </section>
        }

        <!-- PASO 2: DATOS MÉDICOS Y CONTACTO -->
        @if (paso() === 'DATOS_CONTACTO' && ciudadanoValidado(); as ciud) {
          <section class="step-fade">
            <h1 class="step-title">Completa tu registro</h1>

            <form class="form-stack" novalidate (ngSubmit)="completarRegistro()">
              <div class="field">
                <div class="label-between">
                  <label class="field-label" for="nombreOficial">Nombre completo</label>
                  <button type="button" class="change-dni-btn" (click)="reiniciarPaso1()">
                    Cambiar DNI
                  </button>
                </div>
                <div class="locked-wrap">
                  <input
                    id="nombreOficial"
                    name="nombreCompleto"
                    type="text"
                    class="input locked-input"
                    [value]="ciud.nombreCompleto"
                    readonly
                  />
                  <span class="lock-icon" aria-hidden="true">
                    <svg width="17" height="17" viewBox="0 0 24 24" fill="none">
                      <rect x="5" y="11" width="14" height="10" rx="3" stroke="currentColor" stroke-width="2.2" />
                      <path d="M8 11V7.5C8 5.29 9.79 3.5 12 3.5C14.21 3.5 16 5.29 16 7.5V11" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" />
                    </svg>
                  </span>
                </div>
              </div>

              <div class="blood-row">
                <div class="field">
                  <span class="field-label">Tipo de sangre</span>
                  <div class="pill-group">
                    @for (g of grupos; track g) {
                      <button
                        type="button"
                        class="choice-pill tabular"
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
                  <div class="pill-group rh-group">
                    @for (rh of factoresRh; track rh) {
                      <button
                        type="button"
                        class="choice-pill tabular"
                        [class.selected]="factorRh === rh"
                        (click)="factorRh = rh"
                      >
                        {{ rh }}
                      </button>
                    }
                  </div>
                </div>
              </div>

              <div class="two-fields">
                <div class="field">
                  <div class="label-with-tip">
                    <label class="field-label" for="celularInput">Celular</label>

                    <div class="tip-anchor" [class.open]="tooltipCelularAbierto()" (click)="$event.stopPropagation()">
                      <button
                        type="button"
                        class="tip-btn"
                        aria-label="Ver celulares de prueba"
                        [attr.aria-expanded]="tooltipCelularAbierto()"
                        (click)="toggleTooltipCelular($event)"
                      >
                        ?
                      </button>

                      <div class="tip-popover" role="tooltip">
                        <div class="tip-head-row">
                          <span class="tip-head">Celulares de prueba</span>
                          <button type="button" class="tip-refresh" (click)="regenerarCelulares($event)">
                            Otros
                          </button>
                        </div>
                        @for (cel of celularesPrueba(); track cel) {
                          <button
                            type="button"
                            class="tip-option tabular"
                            (click)="seleccionarCelularPrueba(cel)"
                          >
                            {{ cel }}
                          </button>
                        }
                      </div>
                    </div>
                  </div>

                  <input
                    id="celularInput"
                    name="celular"
                    type="tel"
                    inputmode="numeric"
                    maxlength="9"
                    class="input tabular"
                    [class.input-error]="!!errorCelular()"
                    placeholder="987654321"
                    required
                    [ngModel]="celular"
                    (ngModelChange)="onCelularInput($event)"
                    (blur)="validarCelular()"
                  />
                  @if (errorCelular(); as errCel) {
                    <p class="field-error" role="alert">{{ errCel }}</p>
                  }
                </div>

                <div class="field">
                  <div class="label-with-tip">
                    <label class="field-label" for="correoInput">Correo electrónico</label>

                    <div class="tip-anchor" [class.open]="tooltipCorreoAbierto()" (click)="$event.stopPropagation()">
                      <button
                        type="button"
                        class="tip-btn"
                        aria-label="Ver correos de prueba"
                        [attr.aria-expanded]="tooltipCorreoAbierto()"
                        (click)="toggleTooltipCorreo($event)"
                      >
                        ?
                      </button>

                      <div class="tip-popover tip-right" role="tooltip">
                        <div class="tip-head-row">
                          <span class="tip-head">Correos de prueba</span>
                          <button type="button" class="tip-refresh" (click)="regenerarCorreos($event)">
                            Otros
                          </button>
                        </div>
                        @for (mail of correosPrueba(); track mail) {
                          <button
                            type="button"
                            class="tip-option"
                            (click)="seleccionarCorreoPrueba(mail)"
                          >
                            {{ mail }}
                          </button>
                        }
                      </div>
                    </div>
                  </div>

                  <input
                    id="correoInput"
                    name="correo"
                    type="email"
                    autocomplete="email"
                    class="input"
                    [class.input-error]="!!errorCorreo()"
                    placeholder="tu@correo.com"
                    required
                    [(ngModel)]="correo"
                    (ngModelChange)="onCorreoInput($event)"
                    (blur)="validarCorreo()"
                  />
                  @if (errorCorreo(); as errMail) {
                    <p class="field-error" role="alert">{{ errMail }}</p>
                  }
                </div>
              </div>

              <div class="field">
                <span class="field-label">Hospital preferido</span>
                <div class="hospital-grid" role="radiogroup" aria-label="Hospital preferido">
                  @for (h of hospitalOpciones; track h.id) {
                    <button
                      type="button"
                      role="radio"
                      class="hosp-option-card"
                      [class.selected]="bancoPreferidoId === h.id"
                      [attr.aria-checked]="bancoPreferidoId === h.id"
                      (click)="bancoPreferidoId = h.id"
                    >
                      <span class="hosp-card-top">
                        <span class="hosp-card-title">{{ h.nombre }}</span>
                        @if (bancoPreferidoId === h.id) {
                          <svg class="hosp-check" width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                            <path d="M5 13L9.5 17.5L19 7" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round" />
                          </svg>
                        }
                      </span>
                      <span class="hosp-card-sub">{{ h.detalle }}</span>
                    </button>
                  }
                </div>
              </div>

              <div>
                <label class="checkbox-row custom-check">
                  <input
                    type="checkbox"
                    name="aceptaPoliticaPrivacidad"
                    [(ngModel)]="aceptaPoliticaPrivacidad"
                    (ngModelChange)="onPrivacidadChange($event)"
                  />
                  <span>Acepto recibir avisos de donación y el cuidado de mis datos.</span>
                </label>
                @if (errorPrivacidad(); as errPriv) {
                  <p class="field-error" role="alert">{{ errPriv }}</p>
                }
              </div>

              @if (apiError(); as err) {
                <p class="field-error" role="alert">{{ formatErrorText(err) }}</p>
              }

              <button
                type="submit"
                class="btn btn-primary btn-full"
                [disabled]="loading()"
              >
                {{ loading() ? 'Registrando...' : 'Completar registro' }}
              </button>
            </form>
          </section>
        }

        <!-- PANTALLA FINAL: REGISTRO EXITOSO -->
        @if (paso() === 'CONFIRMADO' && donanteRegistrado(); as don) {
          <section class="step-fade success-box">
            <div class="success-icon" aria-hidden="true">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none">
                <path d="M5 13L9.5 17.5L19 7" stroke="#2A8191" stroke-width="2.8" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </div>

            <h1 class="step-title">Registro Exitoso</h1>
            <p class="step-sub">
              Gracias, <strong>{{ don.nombreCompleto }}</strong>. Ya formas parte de la red de
              donantes voluntarios.
            </p>

            <dl class="summary-card tabular">
              <div>
                <dt>Código de donante</dt>
                <dd class="code-mono">{{ don.donanteId }}</dd>
              </div>
              <div>
                <dt>Grupo sanguíneo</dt>
                <dd>{{ don.tipoSangre }}{{ don.factorRh }}</dd>
              </div>
              <div>
                <dt>Tiempo de registro</dt>
                <dd>Completado en {{ segundosTranscurridos() }} segundos</dd>
              </div>
            </dl>

            <a routerLink="/" class="btn btn-secondary btn-full">
              Volver al inicio
            </a>
          </section>
        }
      </main>

      <app-confirm-modal
        [open]="modalSalirAbierto()"
        title="¿Salir sin guardar?"
        message="Si vuelves al inicio ahora, perderás los datos que ingresaste."
        cancelLabel="Seguir aquí"
        confirmLabel="Sí, salir"
        (cancel)="cancelarSalida()"
        (confirm)="confirmarSalida()"
      />
    </div>
  `,
  styles: [`
    .donor-page {
      min-height: 100dvh; background-color: var(--bg-canvas); padding: 2rem 1.25rem;
      display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 1rem;
    }
    .donor-topbar { width: 100%; max-width: 600px; display: flex; align-items: center; justify-content: space-between; padding: 0 0.25rem; }
    .donor-brand { display: inline-flex; align-items: center; gap: 0.5rem; font-size: 1.2rem; font-weight: 900; }
    .brand-dot {
      width: 12px; height: 12px; border-radius: 50%;
      background: linear-gradient(135deg, var(--brand-turquoise) 50%, var(--brand-yellow) 50%);
    }
    .back-link { display: inline-flex; align-items: center; gap: 0.35rem; font-size: 0.94rem; font-weight: 800; color: var(--brand-turquoise-dark); }
    .donor-card {
      width: 100%; max-width: 600px; background-color: var(--bg-surface);
      border: 1px solid var(--border-subtle); border-radius: 32px; padding: 2.25rem;
      box-shadow: var(--shadow-soft); display: flex; flex-direction: column; gap: 1.35rem;
    }
    .progress-row { display: flex; align-items: center; gap: 0.85rem; }
    .progress-track { flex: 1; height: 6px; background-color: #EFE9DF; border-radius: 9999px; overflow: hidden; }
    .progress-fill { height: 100%; background-color: var(--brand-turquoise); border-radius: 9999px; transition: width 320ms cubic-bezier(0.22, 1, 0.36, 1); }
    .progress-label { font-size: 0.8rem; font-weight: 800; color: var(--text-muted); white-space: nowrap; }
    .step-fade { display: flex; flex-direction: column; gap: 1rem; animation: bloodlinkFadeSlide 320ms cubic-bezier(0.22, 1, 0.36, 1) both; }
    .step-title { font-size: 1.75rem; font-weight: 900; }
    .step-sub { font-size: 0.98rem; color: var(--text-secondary); margin-top: -0.35rem; }
    .label-with-tip, .label-between { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; }
    .label-with-tip { justify-content: flex-start; }
    .dni-input { min-height: 54px; font-size: 1.08rem; }
    .btn-full { width: 100%; }
    .change-dni-btn {
      background: none; border: none; font-size: 0.84rem; font-weight: 800;
      color: var(--brand-turquoise-dark); text-decoration: underline; cursor: pointer;
    }
    .locked-wrap { position: relative; display: flex; align-items: center; }
    .locked-input { background-color: var(--brand-turquoise-soft) !important; border-color: transparent !important; font-weight: 700; padding-right: 2.75rem; }
    .lock-icon { position: absolute; right: 1rem; color: var(--brand-turquoise-dark); display: inline-flex; }
    .blood-row, .two-fields { display: grid; grid-template-columns: 1.55fr 1fr; gap: 1rem; }
    .two-fields { grid-template-columns: 1fr 1fr; align-items: start; }
    .pill-group { display: grid; grid-template-columns: repeat(4, 1fr); gap: 0.45rem; }
    .rh-group { grid-template-columns: repeat(2, 1fr); }
    .choice-pill {
      min-height: 50px; border-radius: 9999px; border: 2px solid transparent;
      background-color: var(--bg-surface-muted); font-weight: 700; cursor: pointer; transition: all 200ms ease;
    }
    .choice-pill:hover { background-color: #EAE4D7; }
    .choice-pill.selected {
      background-color: var(--brand-yellow); color: var(--brand-ink); font-weight: 800;
      border-color: var(--brand-yellow-hover); box-shadow: 0 4px 12px rgba(247, 204, 81, 0.35);
    }
    .custom-check { align-items: center; }
    .custom-check input[type="checkbox"] { width: 22px; height: 22px; margin-top: 0; border-radius: 6px; accent-color: var(--brand-turquoise-dark); }
    .success-box { align-items: center; text-align: center; }
    .success-icon {
      width: 58px; height: 58px; border-radius: 50%; background-color: var(--brand-turquoise-soft);
      display: inline-flex; align-items: center; justify-content: center;
    }
    .summary-card {
      width: 100%; padding: 1.25rem; background-color: var(--bg-canvas); border-radius: 20px;
      border: 1px solid var(--border-subtle); display: grid; gap: 0.85rem; text-align: left;
    }
    .summary-card dt { font-size: 0.78rem; font-weight: 700; color: var(--text-muted); }
    .summary-card dd { font-size: 1rem; font-weight: 800; }
    @media (max-width: 560px) {
      .donor-card { padding: 1.5rem; }
      .blood-row, .two-fields { grid-template-columns: 1fr; }
    }
  `],
})
export class RegistroDonanteComponent {
  private readonly donantesService = inject(DonantesService);
  private readonly destroyRef = inject(DestroyRef);
  private readonly router = inject(Router);
  private readonly startedAtMs = Date.now();

  protected readonly hospitalOpciones = HOSPITALES_OPCIONES;
  protected readonly grupos = GRUPOS_SANGUINEOS;
  protected readonly factoresRh = FACTORES_RH;
  protected readonly dnisPrueba = DNIS_PRUEBA;

  protected readonly paso = signal<PasoRegistro>('IDENTIDAD');
  protected readonly loading = signal(false);
  protected readonly apiError = signal<ExtractedError | null>(null);
  protected readonly ciudadanoValidado = signal<CiudadanoReniec | null>(null);
  protected readonly donanteRegistrado = signal<DonanteResponseDTO | null>(null);
  protected readonly segundosTranscurridos = signal<number>(0);
  protected readonly tooltipAbierto = signal(false);
  protected readonly tooltipCelularAbierto = signal(false);
  protected readonly tooltipCorreoAbierto = signal(false);
  protected readonly celularesPrueba = signal<string[]>([]);
  protected readonly correosPrueba = signal<string[]>([]);
  protected readonly modalSalirAbierto = signal(false);

  protected readonly errorCelular = signal('');
  protected readonly errorCorreo = signal('');
  protected readonly errorPrivacidad = signal('');

  private readonly dniSignal = signal('');

  get dni(): string {
    return this.dniSignal();
  }

  set dni(val: string) {
    this.dniSignal.set(val ?? '');
  }

  protected readonly tieneCambiosSinGuardar = computed(
    () =>
      this.paso() === 'DATOS_CONTACTO' ||
      (this.paso() === 'IDENTIDAD' && this.dniSignal().trim().length > 0),
  );

  protected tipoSangre: TipoSangre = 'O';
  protected factorRh: FactorRh = '+';
  protected celular = '';
  protected correo = '';
  protected bancoPreferidoId = 'HOSP-MINSA-01';
  protected aceptaPoliticaPrivacidad = false;

  constructor() {
    this.generarCelularesAleatorios();
    this.generarCorreosAleatorios();
    const intervalId = setInterval(() => {
      this.generarCelularesAleatorios();
      this.generarCorreosAleatorios();
    }, 20000);
    this.destroyRef.onDestroy(() => clearInterval(intervalId));
  }

  toggleTooltipDni(event: Event): void {
    event.stopPropagation();
    this.tooltipCelularAbierto.set(false);
    this.tooltipCorreoAbierto.set(false);
    this.tooltipAbierto.update((v) => !v);
  }

  toggleTooltipCelular(event: Event): void {
    event.stopPropagation();
    this.tooltipAbierto.set(false);
    this.tooltipCorreoAbierto.set(false);
    this.tooltipCelularAbierto.update((v) => !v);
  }

  toggleTooltipCorreo(event: Event): void {
    event.stopPropagation();
    this.tooltipAbierto.set(false);
    this.tooltipCelularAbierto.set(false);
    this.tooltipCorreoAbierto.update((v) => !v);
  }

  cerrarTooltips(): void {
    if (this.tooltipAbierto()) this.tooltipAbierto.set(false);
    if (this.tooltipCelularAbierto()) this.tooltipCelularAbierto.set(false);
    if (this.tooltipCorreoAbierto()) this.tooltipCorreoAbierto.set(false);
  }

  intentarSalir(event?: Event): void {
    event?.preventDefault();
    this.cerrarTooltips();
    if (!this.tieneCambiosSinGuardar()) {
      void this.router.navigateByUrl('/');
      return;
    }
    this.modalSalirAbierto.set(true);
  }

  cancelarSalida(): void {
    this.modalSalirAbierto.set(false);
  }

  confirmarSalida(): void {
    this.modalSalirAbierto.set(false);
    void this.router.navigateByUrl('/');
  }

  seleccionarDniPrueba(dniElegido: string): void {
    this.dni = dniElegido;
    this.tooltipAbierto.set(false);
    this.apiError.set(null);
  }

  regenerarCelulares(event: Event): void {
    event.stopPropagation();
    this.generarCelularesAleatorios();
  }

  seleccionarCelularPrueba(cel: string): void {
    this.celular = cel;
    this.errorCelular.set('');
    this.apiError.set(null);
    this.tooltipCelularAbierto.set(false);
  }

  generarCelularesAleatorios(): void {
    const prefixes = ['98', '96', '94'];
    const generated = prefixes.map((pref) => {
      const rest = Math.floor(1000000 + Math.random() * 9000000);
      return `${pref}${rest}`;
    });
    this.celularesPrueba.set(generated);
  }

  regenerarCorreos(event: Event): void {
    event.stopPropagation();
    this.generarCorreosAleatorios();
  }

  seleccionarCorreoPrueba(email: string): void {
    this.correo = email;
    this.errorCorreo.set('');
    this.apiError.set(null);
    this.tooltipCorreoAbierto.set(false);
  }

  generarCorreosAleatorios(): void {
    const ciud = this.ciudadanoValidado();
    const firstWord = ciud?.nombreCompleto?.trim().split(/\s+/)[0]?.toLowerCase() ?? '';
    const baseName = firstWord.replace(/[^a-z]/g, '') || 'donante';
    const altNames = [baseName, 'piura', 'salud'];
    const domains = ['correo.pe', 'ejemplo.pe', 'donante.pe'];

    const generated = domains.map((dom, idx) => {
      const num = Math.floor(100 + Math.random() * 900);
      return `${altNames[idx]}.${num}@${dom}`;
    });
    this.correosPrueba.set(generated);
  }

  onCelularInput(raw: string): void {
    let digits = (raw || '').replace(/\D/g, '');
    if (digits.startsWith('51') && digits.length >= 11) {
      digits = digits.slice(2);
    }
    this.celular = digits.slice(0, 9);
    if (CELULAR_REGEX.test(this.celular)) {
      this.errorCelular.set('');
    }
    this.apiError.set(null);
  }

  validarCelular(): boolean {
    const clean = this.celular.trim();
    if (!CELULAR_REGEX.test(clean)) {
      this.errorCelular.set('Ingresa un celular de 9 dígitos que empiece con 9.');
      return false;
    }
    this.errorCelular.set('');
    return true;
  }

  onCorreoInput(raw: string): void {
    this.correo = raw;
    if (CORREO_REGEX.test(raw.trim())) {
      this.errorCorreo.set('');
    }
    this.apiError.set(null);
  }

  validarCorreo(): boolean {
    const clean = this.correo.trim();
    if (!CORREO_REGEX.test(clean)) {
      this.errorCorreo.set('Ingresa un correo electrónico válido.');
      return false;
    }
    this.errorCorreo.set('');
    return true;
  }

  onPrivacidadChange(checked: boolean): void {
    if (checked) {
      this.errorPrivacidad.set('');
    }
  }

  validarIdentidad(): void {
    const cleanDni = this.dni.trim();
    if (cleanDni.length !== 8) return;

    this.loading.set(true);
    this.apiError.set(null);
    this.cerrarTooltips();

    this.donantesService.validarDniReniec(cleanDni).subscribe({
      next: (ciudadano) => {
        this.loading.set(false);
        this.ciudadanoValidado.set(ciudadano);
        this.generarCorreosAleatorios();
        this.paso.set('DATOS_CONTACTO');
      },
      error: (err) => {
        this.loading.set(false);
        this.apiError.set(extractApiError(err));
      },
    });
  }

  completarRegistro(): void {
    const ciud = this.ciudadanoValidado();
    if (!ciud) return;

    const celOK = this.validarCelular();
    const mailOK = this.validarCorreo();
    const privOK = this.aceptaPoliticaPrivacidad;
    if (!privOK) {
      this.errorPrivacidad.set('Debes aceptar para continuar.');
    } else {
      this.errorPrivacidad.set('');
    }

    if (!celOK || !mailOK || !privOK) {
      return;
    }

    const payload: RegistrarDonanteRequest = {
      dni: ciud.dni,
      celular: '+51' + this.celular.trim(),
      correo: this.correo.trim(),
      tipoSangre: this.tipoSangre,
      factorRh: this.factorRh,
      bancoPreferidoId: this.bancoPreferidoId,
      aceptaPoliticaPrivacidad: this.aceptaPoliticaPrivacidad,
    };

    this.loading.set(true);
    this.apiError.set(null);
    this.cerrarTooltips();

    this.donantesService.registrarDonante(payload).subscribe({
      next: (donante) => {
        this.loading.set(false);
        const elapsedSec = Math.max(1, Math.round((Date.now() - this.startedAtMs) / 1000));
        this.segundosTranscurridos.set(elapsedSec);
        this.donanteRegistrado.set(donante);
        this.paso.set('CONFIRMADO');
      },
      error: (err) => {
        this.loading.set(false);
        this.apiError.set(extractApiError(err));
      },
    });
  }

  reiniciarPaso1(): void {
    this.paso.set('IDENTIDAD');
    this.ciudadanoValidado.set(null);
    this.apiError.set(null);
    this.errorCelular.set('');
    this.errorCorreo.set('');
    this.errorPrivacidad.set('');
    this.cerrarTooltips();
  }

  formatErrorText(err: ExtractedError): string {
    const raw = (err.mensaje || err.detalle || '').trim();
    if (!raw) return 'No pudimos completar la validación. Verifica los datos ingresados.';
    const capitalized = raw.charAt(0).toUpperCase() + raw.slice(1);
    return /[.!?]$/.test(capitalized) ? capitalized : `${capitalized}.`;
  }

  nuevoRegistro(): void {
    this.dni = '';
    this.celular = '';
    this.correo = '';
    this.aceptaPoliticaPrivacidad = false;
    this.ciudadanoValidado.set(null);
    this.donanteRegistrado.set(null);
    this.apiError.set(null);
    this.errorCelular.set('');
    this.errorCorreo.set('');
    this.errorPrivacidad.set('');
    this.paso.set('IDENTIDAD');
  }
}
