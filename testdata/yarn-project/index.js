import express from 'express';
import { merge } from 'lodash';

const app = express();
const config = merge({ a: 1 }, { b: 2 });

app.listen(3000);
export { app, config };
