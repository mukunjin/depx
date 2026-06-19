import axios from 'axios';
import { debounce } from 'lodash';

const client = axios.create();
const fn = debounce(() => {}, 300);
