import axios from 'axios';
import { debounce } from 'lodash';

async function fetchData() {
  const response = await axios.get('https://api.example.com/data');
  return response.data;
}

const debouncedFetch = debounce(fetchData, 300);
