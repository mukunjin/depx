import express from 'express';
import lodash from 'lodash';

const app = express();
const data = lodash.cloneDeep({ test: 1 });

app.get('/', (req, res) => {
  res.json(data);
});
