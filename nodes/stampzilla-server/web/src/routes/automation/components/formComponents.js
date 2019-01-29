import PropTypes from 'prop-types';
import React from 'react';
import classnames from 'classnames';

export const ObjectFieldTemplate = (props) => {
  if (props.title !== undefined) {
    return (
      <div className="card mb-4 bg-dark text-white">
        <div className="card-header">
          <div><strong>{props.title}</strong></div>
          <small>{props.description}</small>
        </div>
        <div className={classnames({
              'card-body': true,
          })}
        >
          {props.properties.map(prop => prop.content)}
        </div>
      </div>
    );
  }

  return typeof props.uiSchema.tableMode !== 'undefined' ? (
    <div className="row">
      {props.properties.map(prop => <div key={prop.name} className="col-sm-6">{prop.content}</div>)}
    </div>
  ) : (
    <React.Fragment>
      {props.properties.map(prop => prop.content)}
    </React.Fragment>
  );
};
ObjectFieldTemplate.propTypes = {
  title: PropTypes.string,
  description: PropTypes.string,
  properties: PropTypes.arrayOf(PropTypes.shape({
    content: PropTypes.node,
  })),
  uiSchema: PropTypes.shape({
    tableMode: PropTypes.bool,
  }),
};

export const CustomCheckbox = (props) => {
  const {
    id,
    value,
    required,
    disabled,
    readonly,
    label,
    autofocus,
    onChange,
  } = props;
  return (
    <div className={`checkbox custom-control custom-checkbox ${disabled || readonly ? 'disabled' : ''}`}>
      <input
        type="checkbox"
        className="custom-control-input"
        id={id}
        checked={typeof value === 'undefined' ? false : value}
        required={required}
        disabled={disabled || readonly}
        autoFocus={autofocus}
        onChange={event => onChange(event.target.checked)}
      />
      <label className="custom-control-label" htmlFor={id}>
        <span>{label}</span>
      </label>
    </div>
  );
};
CustomCheckbox.propTypes = {
  id: PropTypes.string,
  value: PropTypes.bool,
  required: PropTypes.bool,
  disabled: PropTypes.bool,
  readonly: PropTypes.bool,
  label: PropTypes.string,
  autofocus: PropTypes.bool,
  onChange: PropTypes.func,
};


const IconButton = (props) => {
  const {
    type = 'default', icon, className, ...otherProps
  } = props;
  return (
    <button
      type="button"
      className={`btn btn-${type} ${className}`}
      {...otherProps}
    >
      <i className={`fa fa-${icon}`} />
    </button>
  );
};
IconButton.propTypes = {
  type: PropTypes.string,
  icon: PropTypes.string,
  className: PropTypes.string,
};

const DefaultArrayItem = (props) => {
  const btnStyle = {
    flex: 1,
    paddingLeft: 6,
    paddingRight: 6,
    fontWeight: 'bold',
  };
  return (
    <div key={props.index} className={classnames(['row mb-3', props.className])}>
      <div className={props.hasToolbar ? 'col-sm-9' : 'col-sm-12'}>
        {props.children}
      </div>

      {props.hasToolbar && (
        <div className="col-sm-3 array-item-toolbox">
          <div
            className="btn-group"
            style={{
            display: 'flex',
              justifyContent: 'space-around',
            }}
          >
            {(props.hasMoveUp || props.hasMoveDown) && (
              <IconButton
                type="secondary"
                icon="arrow-up"
                className="array-item-move-up"
                tabIndex="-1"
                style={btnStyle}
                disabled={props.disabled || props.readonly || !props.hasMoveUp}
                onClick={props.onReorderClick(props.index, props.index - 1)}
              />
            )}
            {(props.hasMoveUp || props.hasMoveDown) && (
              <IconButton
                type="secondary"
                icon="arrow-down"
                className="array-item-move-down"
                tabIndex="-1"
                style={btnStyle}
                disabled={
                  props.disabled || props.readonly || !props.hasMoveDown
                }
                onClick={props.onReorderClick(props.index, props.index + 1)}
              />
            )}
            {props.hasRemove && (
              <IconButton
                type="danger"
                icon="trash"
                className="array-item-remove"
                tabIndex="-1"
                style={btnStyle}
                disabled={props.disabled || props.readonly}
                onClick={props.onDropIndexClick(props.index)}
              />
            )}
          </div>
        </div>
      )}
    </div>
  );
};

DefaultArrayItem.propTypes = {
  // schema: PropTypes.object.isRequired,
  uiSchema: PropTypes.shape({
    'ui:options': PropTypes.shape({
      addable: PropTypes.bool,
      orderable: PropTypes.bool,
      removable: PropTypes.bool,
    }),
  }),
  // idSchema: PropTypes.object,
  // errorSchema: PropTypes.object,
  // onChange: PropTypes.func.isRequired,
  // onBlur: PropTypes.func,
  // onFocus: PropTypes.func,
  // formData: PropTypes.array,
  // required: PropTypes.bool,
  disabled: PropTypes.bool,
  readonly: PropTypes.bool,
  // autofocus: PropTypes.bool,
  registry: PropTypes.shape({
    widgets: PropTypes.objectOf(
      PropTypes.oneOfType([PropTypes.func, PropTypes.shape({})]),
    ).isRequired,
    fields: PropTypes.objectOf(PropTypes.func).isRequired,
    definitions: PropTypes.object.isRequired,
    formContext: PropTypes.object.isRequired,
  }),
  onDropIndexClick: PropTypes.func,
  hasRemove: PropTypes.bool,
  hasMoveUp: PropTypes.bool,
  hasMoveDown: PropTypes.bool,
  children: PropTypes.node,
  hasToolbar: PropTypes.bool,
  className: PropTypes.string,
  index: PropTypes.string,
  onReorderClick: PropTypes.func,
};

export const ArrayFieldTemplate = (props) => {
  const { title } = props;
  return (
    <div>
      <label>
        <span>{title}</span>
      </label>
      {props.items && props.items.map(DefaultArrayItem)}
      {props.canAdd && <IconButton icon="plus" className="btn-block" onClick={props.onAddClick} />}
    </div>
  );
};
ArrayFieldTemplate.propTypes = {
  title: PropTypes.string,
  canAdd: PropTypes.bool,
  onAddClick: PropTypes.func,
  items: PropTypes.arrayOf(PropTypes.shape({})),
};
