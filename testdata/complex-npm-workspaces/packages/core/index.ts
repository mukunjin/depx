import lodash from 'lodash';
import moment from 'moment';

export const formatDate = (date: Date) => moment(date).format('YYYY-MM-DD');
export const clone = (obj: any) => lodash.cloneDeep(obj);
