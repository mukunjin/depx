import axios from 'axios';
import chalk from 'chalk';

export const fetchData = async (url: string) => {
  console.log(chalk.blue('Fetching data...'));
  return axios.get(url);
};
