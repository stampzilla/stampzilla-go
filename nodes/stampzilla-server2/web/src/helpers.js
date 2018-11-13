
let lastId = 0;
export const uniqeId = (prefix='id') =>  {
    lastId++;
    return `${prefix}${lastId}`;
};

export default false;
