import {
  Button,
  Modal,
  ModalBody,
  ModalFooter,
  ModalHeader,
} from 'reactstrap';
import Form from 'react-jsonschema-form';
import React from 'react';

import styles from './FormModal.scss';

const log = type => console.log.bind(console, type); // eslint-disable-line no-console

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

class FormModal extends React.Component {
  state = {
    isValid: true,
    formData: {
    },
  }

  componentWillReceiveProps(props) {
    this.setState({
      formData: props.formData,
    });
  }

  onChange() {
    return (data) => {
      const { errors, formData } = data;
      this.setState({
        isValid: errors.length === 0,
        formData,
      });

      if (this.props.onChange) {
        this.props.onChange(data);
      }
    };
  }

  render() {
    return (
      <Modal isOpen={this.props.isOpen} toggle={this.props.onClose} wrapClassName="bootstrap-global" className={styles.formModal}>
        {this.props.title &&
        <ModalHeader toggle={this.props.onClose}>
          {this.props.title}
        </ModalHeader>
        }
        <ModalBody>
          {!this.props.prepend && this.props.children}
          <Form
            schema={this.props.schema}
            uiSchema={this.props.uiSchema}
            showErrorList={false}
            liveValidate
            onChange={this.onChange()}
            formData={this.state.formData}
            onSubmit={this.props.onSubmit}
            onError={log('errors')}
            disabled={this.props.disabled}
            transformErrors={this.props.transformErrors}
            widgets={{
              CheckboxWidget: CustomCheckbox,
            }}
          >
            <button ref={(btn) => { this.submitButton = btn; }} style={{ display: 'none' }} type="submit" />
          </Form>
          {this.props.prepend && this.props.children}
        </ModalBody>
        <ModalFooter>
          {this.props.footer}
          <Button color="primary" disabled={!this.state.isValid || this.props.disabled} onClick={() => this.submitButton.click()}>
            {this.props.submitButton || 'Create'}
          </Button>{' '}
        </ModalFooter>
      </Modal>
    );
  }
}

export default FormModal;
