export function formatRol(rol?: string | null): string {
  switch (rol) {
    case 'MEDICO':
      return 'Médico';
    case 'PERSONAL_BANCO':
      return 'Personal de Banco';
    case 'COORDINADOR':
      return 'Coordinador';
    default:
      return rol || 'Usuario';
  }
}

export function formatUrgencia(urgencia?: string | null): string {
  switch (urgencia) {
    case 'ALTA_VITAL':
      return 'Emergencia vital';
    case 'MEDIA':
      return 'Prioridad media';
    default:
      return urgencia || 'Prioridad estándar';
  }
}

export function formatOrigenAsignacion(origen?: string | null): string {
  switch (origen) {
    case 'LOCAL':
      return 'Tu mismo hospital';
    case 'RED_EXTERNA':
      return 'Banco cercano de la red';
    case 'NINGUNO':
      return 'Sin stock disponible';
    default:
      return origen || 'En evaluación';
  }
}

export function formatEstadoSolicitud(estado?: string | null): string {
  switch (estado) {
    case 'En Espera de Donantes':
      return 'Buscando donantes';
    default:
      return estado || 'Registrada';
  }
}

export function formatAlertaCaducidad(alerta?: string | null): string {
  switch (alerta) {
    case 'VERDE':
      return 'Vigente';
    case 'AMARILLO':
      return 'Próxima a vencer';
    case 'ROJO':
      return 'Vencimiento crítico';
    default:
      return alerta || 'Vigente';
  }
}

export function formatNivelAbastecimiento(nivel?: string | null): string {
  switch (nivel) {
    case 'ESTABLE':
      return 'Nivel saludable';
    case 'CRITICO':
      return 'Stock bajo';
    default:
      return nivel || 'Nivel saludable';
  }
}

export function formatOrigenUnidad(origen?: string | null): string {
  switch (origen) {
    case 'DONACION_VOLUNTARIA':
      return 'Donación voluntaria';
    case 'CAMPANA_REGIONAL_MINSA':
      return 'Campaña regional';
    case 'REPOSICION_FAMILIAR':
      return 'Reposición familiar';
    default:
      return origen || 'Donación voluntaria';
  }
}

export function formatEventoNotificacion(evento?: string | null): string {
  switch (evento) {
    case 'UnidadRegistrada':
      return 'Nueva unidad registrada';
    case 'EstadoUnidadActualizado':
      return 'Cambio de estado de unidad';
    case 'UnidadProximaACaducar':
      return 'Unidades próximas a vencer';
    case 'EmergenciaReservada':
      return 'Reserva de emergencia confirmada';
    case 'SolicitudCancelada':
      return 'Solicitud cancelada';
    case 'StockCriticoDetectado':
      return 'Convocatoria de donantes enviada';
    default:
      return evento || 'Aviso del sistema';
  }
}

export function formatCanalNotificacion(canal?: string | null): string {
  switch (canal) {
    case 'ALERTA_PANEL':
      return 'Aviso en sistema';
    case 'SMS':
      return 'Mensaje SMS';
    case 'EMAIL':
      return 'Correo electrónico';
    default:
      return canal || 'Aviso en sistema';
  }
}

export function formatAccionAuditoria(accion?: string | null): string {
  switch (accion) {
    case 'REGISTRO_INICIAL':
      return 'Ingreso inicial al banco';
    case 'REGISTRAR_INGRESO_UNIDAD':
      return 'Registro de nueva unidad';
    case 'ACTUALIZAR_ESTADO_UNIDAD':
      return 'Actualización de estado';
    case 'RESERVA_AUTOMATICA_MATCHING':
      return 'Reserva para emergencia';
    case 'RESERVA_SOAP_LEGACY':
      return 'Reserva en hospital conectado';
    case 'CONFIRMAR_TRANSFUSION':
      return 'Transfusión realizada';
    case 'LIBERAR_RESERVA_POR_CANCELACION':
      return 'Unidad liberada por cancelación';
    default:
      return accion || 'Movimiento registrado';
  }
}
