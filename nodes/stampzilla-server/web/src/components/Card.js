import React from 'react';
import classnames from 'classnames';

const Card = (props) => {
  const {
    title, children, toolbar, bodyClassName, className,
  } = props;
  return (
    <div className={classnames('card', className)}>
      <div className="card-header no-border">
        <h3 className="card-title" dangerouslySetInnerHTML={{ __html: title }} />
        <div className="card-tools">
          {toolbar && toolbar.map(tool => (
            <button onClick={tool.onClick} className={classnames('btn btn-tool btn-sm', tool.className)} key={tool.icon}>
              <i className={tool.icon} />
            </button>
          ))}
        </div>
      </div>
      <div className={classnames(['card-body', bodyClassName])}>
        {children}
      </div>
    </div>
  );
};

export default Card;
