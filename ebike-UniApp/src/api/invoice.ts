import { request } from '@/shared/request'

export function getInvoiceList(data: Record<string, unknown>) {
  return request({ url: '/client/invoice/page', method: 'POST', data })
}

export function createInvoice(data: Record<string, unknown>) {
  return request({ url: '/client/invoice/create', method: 'POST', data })
}

export function getInvoicedOrders(data: Record<string, unknown>) {
  return request({ url: '/client/order/listInvoicedOrders', method: 'POST', data })
}
