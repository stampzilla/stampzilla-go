import React, { Component } from 'react';
import {
  Button,
} from 'reactstrap';
import { connect } from 'react-redux';
import Form from 'react-jsonschema-form';
import JSONInput from 'react-json-editor-ajrm';
import locale from 'react-json-editor-ajrm/locale/en';

import { write } from '../../components/Websocket';
import Card from '../../components/Card';

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

const JsonWidget = (props) => {
  const {
    value,
    onChange,
  } = props;

  let parsedValue = {};
  try {
    parsedValue = value && JSON.parse(value);
  } catch (err) {
    parsedValue = {
      'parse error': err,
    };
  }

  return (
    <JSONInput
      placeholder={typeof parsedValue === 'object' ? parsedValue : undefined}
      onChange={({ jsObject }) => onChange(jsObject && JSON.stringify(jsObject))}
      theme="dark_vscode_tribute"
      locale={locale}
      height="550px"
      width="100%"
      reset={false}
      waitAfterKeyPress={60000}
    />
  );
};

const schema = {
  type: 'object',
  required: [
    'name',
    'config',
  ],
  properties: {
    name: {
      type: 'string',
      title: 'Name',
    },
    config: {
      type: 'string',
      title: 'Config',
    },
  },
};
const uiSchema = {
  config: {
    'ui:widget': 'json',
    'ui:options': {
      rows: 15,
    },
  },
};

class Node extends Component {
  state = {
    isValid: true,
    formData: {
    },
  }

  componentDidMount = () => {
    this.componentWillReceiveProps(this.props);
  }

  componentWillReceiveProps = (props) => {
    const { nodes, match } = props;
    const node = nodes && nodes.find(n => n.get('uuid') === match.params.uuid);
    if (node) {
      this.setState({
        formData: {
          name: node.get('name'),
          config: JSON.stringify(node.get('config')),
          //config: node.get('config'),
        },
      });
    }
  }

  onChange = () => (data) => {
    const { errors, formData } = data;
    this.setState({
      isValid: errors.length === 0,
      formData,
    });
  }

  onSubmit = () => ({ formData }) => {
    const { nodes, match } = this.props;
    const node = nodes.find(n => n.get('uuid') === match.params.uuid);

    write({
      type: 'setup-node',
      body: {
        ...node.toJS(),
        ...formData,
        config: JSON.parse(formData.config),
      },
    });
  }

  onClickNode = uuid => () => {
    const { history } = this.props;
    history.push(`/nodes/${uuid}`);
  }

  render() {
    const { nodes, match } = this.props;
    const node = nodes.find(n => n.get('uuid') === match.params.uuid);

    return (
      <React.Fragment>
        <div className="row">
          <div className="col-md-12">
            <Card
              title={node ? `Settings for node <strong>${node.get('uuid')}</strong> (type <strong>${node.get('type')}</strong>)` : 'Settings'}
              bodyClassName="p-0"
            >
              <div className="card-body">
                <Form
                  schema={schema}
                  uiSchema={uiSchema}
                  showErrorList={false}
                  liveValidate
                  onChange={this.onChange()}
                  formData={this.state.formData}
                  onSubmit={this.onSubmit()}
                  // onError={log('errors')}
                  // disabled={this.props.disabled}
                  // transformErrors={this.props.transformErrors}
                  widgets={{
                    CheckboxWidget: CustomCheckbox,
                    json: JsonWidget,
                  }}
                >
                  <button ref={(btn) => { this.submitButton = btn; }} style={{ display: 'none' }} type="submit" />
                </Form>
              </div>
              <div className="card-footer">
                <Button color="primary" disabled={!this.state.isValid || this.props.disabled} onClick={() => this.submitButton.click()}>
                  {'Save'}
                </Button>
              </div>
            </Card>
          </div>
        </div>
      </React.Fragment>
    );
  }
}

const mapToProps = state => ({
  nodes: state.getIn(['nodes', 'list']),
  connections: state.getIn(['connections', 'list']),
});

export default connect(mapToProps)(Node);
