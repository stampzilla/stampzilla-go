import React from 'react';

const CustomCheckbox = (props) => {
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

export default CustomCheckbox;
