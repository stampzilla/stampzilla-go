
let lastId = 0;
export const uniqeId = (prefix = 'id') => {
  lastId += 1;
  return `${prefix}${lastId}`;
};

export default false;
