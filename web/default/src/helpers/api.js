import { showError } from './utils';
import axios from 'axios';

export const API = axios.create({
  baseURL: process.env.REACT_APP_SERVER ? process.env.REACT_APP_SERVER : '',
  withCredentials: true,
});

API.interceptors.response.use(
  (response) => response,
  (error) => {
    showError(error);
  }
);
