export interface Notificacion {
  id: number;
  evento: string;
  canal: 'SMS' | 'EMAIL' | 'ALERTA_PANEL';
  destinatario: string;
  mensaje: string;
  timestamp: string;
}

export interface UnidadPorCaducar {
  unidadId: string;
  codigoBarras: string;
  bancoId: string;
  tipoSangre: string;
  factorRh: string;
  fechaCaducidad: string;
  diasRestantes: number;
  alertaCaducidad: 'AMARILLO' | 'ROJO';
}

export interface VerificacionCaducidadResponse {
  totalDetectadas: number;
  alertasPublicadas: number;
  unidades: UnidadPorCaducar[];
  ejecutadoEn: string;
}
