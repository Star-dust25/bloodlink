import { FactorRh, TipoSangre } from './inventario.model';

export interface ValidarDNIRequest {
  dni: string;
}

export interface CiudadanoReniec {
  dni: string;
  dniValido: boolean;
  nombreCompleto: string;
  fechaNacimiento: string;
  mayorEdad: boolean;
}

export interface RegistrarDonanteRequest {
  dni: string;
  celular: string;
  correo: string;
  tipoSangre: TipoSangre;
  factorRh: FactorRh;
  bancoPreferidoId: string;
  aceptaPoliticaPrivacidad: boolean;
}

export interface DonanteResponseDTO {
  donanteId: string;
  dni: string;
  nombreCompleto: string;
  celular?: string;
  correo?: string;
  tipoSangre: TipoSangre;
  factorRh: FactorRh;
  bancoPreferidoId: string;
  validadoReniec: boolean;
  habilitadoConvocatorias: boolean;
  fechaRegistro: string;
}
