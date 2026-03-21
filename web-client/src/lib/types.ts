export type BloodType = 'A+' | 'A-' | 'B+' | 'B-' | 'AB+' | 'AB-' | 'O+' | 'O-';
export type RequestStatus = 'Pending' | 'Processing' | 'Completed' | 'Failed' | 'Canceled' | 'Fulfilled';
export type ActionStatus = 'Pending' | 'Accepted' | 'Declined' | 'Donated';

export interface User {
  id: string;
  name: string;
  email: string;
  phone: string;
  health?: { info_type: string; details: string }[];
  created_at: string;
}

export interface DonationRequest {
  id: string;
  user_id: string;
  location_hex: string;
  location_lat: number;
  location_lng: number;
  bag_count: number;
  required_by_date: string;
  blood_type: BloodType;
  description: string;
  requester_info: string;
  location_name: string;
  status: RequestStatus;
  created_at: string;
  updated_at: string;
}

export interface ActionedUser {
  user_id: string;
  lat: number;
  lng: number;
  h3_hex: string;
  action: ActionStatus;
}

export interface ExtendedDonationRequest {
  request: DonationRequest;
  notified_users: ActionedUser[];
}

export interface ListRequestsResponse {
  requests: DonationRequest[];
  last_request_id?: string;
  page_size: number;
}

export interface Notification {
  id: string;
  type: string;
  title: string;
  content: string;
  metadata?: any;
  created_at: string;
}

export interface NotificationsResponse {
  notifications: Notification[];
  last_notification_id?: string;
  page_size: number;
}

export interface LoginResponse {
  id: string;
  token: string;
}

export interface SignupRequest {
  name: string;
  email: string;
  phone: string;
  password: string;
}
