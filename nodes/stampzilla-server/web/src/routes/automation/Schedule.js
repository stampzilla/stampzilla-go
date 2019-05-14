import React, { Component } from 'react';
import { Button } from 'reactstrap';
import { connect } from 'react-redux';
import Form from 'react-jsonschema-form';

import { add, save } from '../../ducks/schedules';
import Card from '../../components/Card';
import SavedStateWidget from './components/SavedStatePicker';
import {
  ArrayFieldTemplate,
  CustomCheckbox,
  ObjectFieldTemplate,
} from './components/formComponents';

const schema = {
  type: 'object',
  required: ['name'],
  properties: {
    name: {
      type: 'string',
      title: 'Name',
    },
    enabled: {
      type: 'boolean',
      title: 'Enabled',
      description: 'Turn on and off this schedule',
    },
    when: {
      type: 'string',
      title: 'When',
      description: 'Cron formated when line [second minute hour day month dow]',
    },
    actions: {
      type: 'array',
      title: 'Actions',
      items: {
        type: 'string',
      },
    },
  },
};
const uiSchema = {
  config: {
    'ui:options': {
      rows: 15,
    },
  },
  actions: {
    items: {
      'ui:widget': 'SavedStateWidget',
      'ui:options': {
        hideDelay: true,
      },
    },
  },
};

class Automation extends Component {
  constructor(props) {
    super();

    const { schedules, match } = props;
    const schedule = schedules.find(n => n.get('uuid') === match.params.uuid);
    const formData = schedule && schedule.toJS();
    if (formData && formData.actions == null) {
      formData.actions = [];
    }
    this.state = {
      formData,
      isValid: true,
    };
  }

  componentWillReceiveProps(nextProps) {
    const { schedules, match } = nextProps;
    if (
      !this.props
      || match.params.uuid !== this.props.match.params.uuid
      || schedules !== this.props.schedules
    ) {
      const schedule = schedules.find(n => n.get('uuid') === match.params.uuid);
      const formData = schedule && schedule.toJS();
      if (formData && formData.actions == null) {
        formData.actions = [];
      }
      this.setState({
        formData,
      });
    }
  }

  onChange = () => (data) => {
    const { errors, formData } = data;
    this.setState({
      isValid: errors.length === 0,
      formData,
    });
  };

  onSubmit = () => ({ formData }) => {
    const { dispatch } = this.props;

    if (formData.uuid) {
      dispatch(save(formData));
    } else {
      dispatch(add(formData));
    }

    const { history } = this.props;
    history.push('/aut');
  };

  render() {
    const { match } = this.props;

    return (
      <React.Fragment>
        <div className="row">
          <div className="col-md-12">
            <Card
              title={match.params.uuid ? 'Edit schedule ' : 'New schedule'}
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
                  ObjectFieldTemplate={ObjectFieldTemplate}
                  ArrayFieldTemplate={ArrayFieldTemplate}
                  widgets={{
                    CheckboxWidget: CustomCheckbox,
                    SavedStateWidget,
                  }}
                >
                  <button
                    ref={(btn) => {
                      this.submitButton = btn;
                    }}
                    style={{ display: 'none' }}
                    type="submit"
                  />
                </Form>
              </div>
              <div className="card-footer">
                <Button
                  color="primary"
                  disabled={!this.state.isValid || this.props.disabled}
                  onClick={() => this.submitButton.click()}
                >
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
  schedules: state.getIn(['schedules', 'list']),
});

export default connect(mapToProps)(Automation);
