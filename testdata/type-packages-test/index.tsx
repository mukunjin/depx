import express from 'express';
import React from 'react';
import { createRoot } from 'react-dom/client';

const app = express();
app.listen(3000);

const root = createRoot(document.getElementById('root'));
root.render(<div>Hello</div>);
