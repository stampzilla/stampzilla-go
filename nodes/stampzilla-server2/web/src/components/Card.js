import React, { Component } from "react";
import classnames from 'classnames';

const Card = (props) => {
  const { title, children, toolbar, bodyClassName } = props;
  return (
    <div className="card">
      <div className="card-header no-border">
        <h3 className="card-title">{title}</h3>
        <div className="card-tools">
          {toolbar && toolbar.map(tool => (
          <a onClick={tool.onClick} className="btn btn-tool btn-sm">
            <i className={tool.icon}></i>
          </a>
          ))}
        </div>
      </div>
      <div className={classnames(["card-body", bodyClassName])}>
        {children}
      </div>
    </div>
  );
}

export default Card;
