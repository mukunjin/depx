import axios from 'axios';
import { debounce } from 'lodash';

const api = axios.create({ baseURL: 'https://api.example.com' });
const debouncedFn = debounce(() => console.log('debounced'), 300);

export { api, debouncedFn };
