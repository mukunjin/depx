import express from 'express';
import { debounce } from 'lodash';

const app = express();
const debouncedFn = debounce(() => console.log('debounced'), 300);

export { app, debouncedFn };
