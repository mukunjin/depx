import express from 'express';
import React from 'react';
import { createRoot } from 'react-dom/client';
import axios from 'axios';

const app = express();
const root = createRoot(document.getElementById('root'));
const api = axios.create({ baseURL: 'https://api.example.com' });

root.render(<div>Hello</div>);

export { app, api };
