var AddForm = React.createClass({

  getInitialState: function() {
    return {
      clicked: false,
      errored: false,
      json_workload_checked: false,
      cron_check: {
        status: null,
        message: ""
      }
    };
  },

  raiseJSONWorkloadWarning: function(workload) {
    if (workload.length === 0 || workload[0] != "{") {
      return false;
    }

    try {
      JSON.parse(workload);
    } catch(err) {
      return true;
    }

    return false;
  },

  formSubmit: function(e) {
    e.preventDefault();
    var crontime = this.refs.crontime.getInputDOMNode().value;
    var workload = this.refs.workload.getInputDOMNode().value.trim();
    var backend = this.refs.backend.getValue();

    if (!this.state.json_workload_checked && this.raiseJSONWorkloadWarning(workload)) {
      this.setState({json_workload_checked:true});
      return;
    }

    $.ajax({
      url: "/jobs",
      type: "POST",
      data: {Function: this.props.function, CronTime: crontime, Workload: workload, Backend: backend},
      dataType: "json",
      success: function(data) {
        this.props.getJobDetails(this.props.function);
      }.bind(this),
      error: function(xhr, status, err) {
        this.setState({errored: true, err_msg: xhr.responseText});
      }.bind(this)
    });
  },

  addJob: function() {
    this.setState({clicked: true});
  },

  validateCron: function(evt) {
    var crontime = this.refs.crontime.getInputDOMNode().value;
    var cron_check = { };
    try {
      cron_check.message = cronstrue.toString(crontime);
      cron_check.status = "success";
    } catch(e) {
      cron_check.status = "error";
      cron_check.message = e;
    }
    return this.setState({cron_check: cron_check});
  },

  cronAlert: function() {
    if (!this.state.errored) {
      return;
    }
    return <Alert bsStyle="danger">{this.state.err_msg}</Alert>;
  },

  workloadAlert: function() {
    if (this.state.json_workload_checked) {
      var msg = "Workload input looks like JSON but doesn't parse correctly. Submit anyways?";
      return <Alert bsStyle="warning">{msg}</Alert>;
    }
  },

  render: function() {
    if (!this.state.clicked) {
      return (<div><Button bsStyle="primary" onClick={this.addJob}>Add job</Button><p></p></div>);
    }

    var crontime_placeholder = 'e.g.  0 13 */4 * * 1-5 -> run at 0 seconds, 13 mins, every hour divisable by 4, every day and every month on Monday though Friday';
    var workload_placeholder = 'e.g. {"district_id":"12345"}';
    var backend_placeholder = 'e.g. gearman or workflow-mananger';
    var button_msg = this.state.json_workload_checked ? "Yes, submit":"Submit job";
    return (
      <div>
        {this.cronAlert()}
        {this.workloadAlert()}
        <form onSubmit={this.formSubmit} method="POST">
        <label>Cron Time</label>
        <Input hasFeedback help={this.state.cron_check.message} bsStyle={this.state.cron_check.status}
          onChange={this.validateCron} ref="crontime" type="text" placeholder={crontime_placeholder} required />
        <label>Workload</label>
        <Input ref="workload" type="text" placeholder={workload_placeholder} />
        <label>Backend</label>
        <Input type="select" ref="backend">
          <option value="workflow-manager">workflow-manager</option>
          <option value="gearman">gearman</option>
        </Input>
        <ButtonInput bsStyle="primary" type="submit">{button_msg}</ButtonInput>
        </form>
      </div>
    );
  }
});

var CronRow = React.createClass({
  getInitialState: function() {
    return {button_clicked: false, delete_clicked: false};
  },

  toggle_activated_click: function() {
    if (!this.state.activated_clicked) {
      this.setState({activated_clicked: true});
      return;
    }

    var job = this.props.job;
    $.ajax({
      url: "/jobs/" + job.ID,
      type: "PUT",
      data: {IsActive: !job.IsActive,
             Function: job.Function,
             CronTime: job.CronTime,
             Workload: job.Workload,
             Created: job.Created,
             TimeZone: job.TimeZone,
             Backend: job.Backend},
      dataType: "json",
      success: function(data) {
        this.props.getJobDetails(this.props.job.Function);
      }.bind(this),
      error: function(xhr, status, err) {
        console.log("Error " + xhr.responseText);
      }.bind(this)
    });
  },

  delete_click: function() {
    if (!this.state.delete_clicked) {
      this.setState({delete_clicked: true});
      return;
    }

    var job = this.props.job;
    $.ajax({
      url: "/jobs/" + job.ID,
      type: "DELETE",
      success: function(data) {
        this.props.getJobDetails(this.props.job.Function);
      }.bind(this),
      error: function(xhr, status, err) {
        console.log("Error " + xhr.responseText);
      }.bind(this)
    });
  },

  formatTime: function(created) {
    input_format = "YYYY-MM-DDTHH:mm:SSSSZ";
    output_format = "YYYY-MM-DD";
    return moment(created, input_format).format(output_format);
  },

  render: function() {
    var job = this.props.job;
    var display = this.props.job.IsActive ? "Deactivate":"Activate";
    var button_display = this.state.activated_clicked ? "Are you sure?":display;
    var style = this.props.job.IsActive ? "danger":"warning";
    var delete_display = this.state.delete_clicked ? "Are you sure?":"Delete Job";
    var cronString = cronstrue.toString(job.CronTime);
    let workload;
    try {
      var parsedObj = JSON.parse(job.Workload);
      var parsedStr = JSON.stringify(parsedObj, null, 2);
      workload = (<div><pre>{parsedStr}</pre></div>);
    } catch (e) {
      // fallback
      workload = job.Workload;
    }
    return(
      <tr>
        <td id="button"><Button bsStyle={style} onClick={this.toggle_activated_click}>{button_display}</Button></td>
        <td>
          {job.CronTime}
          <hr></hr>
          {cronString}
        </td>
        <td id="workload">{workload}</td>
        <td>{this.formatTime(job.Created)}</td>
        <td>{job.Backend}</td>
        <td id="button"><Button bsStyle="danger" onClick={this.delete_click}>{delete_display}</Button></td>
      </tr>
    );
  }
});

var JobDetails = React.createClass({

  getInitialState: function() {
    var urlFunction = null;
    if (this.props.urlParams && this.props.urlParams.length >= 1) {
      urlFunction = this.props.urlParams[0];
    }

    var func = (urlFunction || this.props.function);
    history.replaceState(null, null, "#jobdetails/" + func);

    this.getJobDetails(func);
    return {loading: true, jobs: [], function: func};
  },

  getJobDetails: function(func) {
    this.setState({loading: true});
    $.ajax({
      type: "GET",
      url: "/jobs",
      data: {"Function": func},
      success: function(data) {
        this.setState({loading: false, jobs: data});
      }.bind(this),
      error : function(xhr, status, err) {
        this.setState({loading: false});
        console.log("Got api error :" + xhr.responseText);
      }.bind(this)
    });
  },

  sortJobs: function(jobs) {
    return jobs.sort(function(a, b) {
      if (a.Workload === b.Workload) {
        return 0;
      }
      return (b.Workload < a.Workload ? 1 : -1);
    });
  },

  displayRows: function(title, jobs) {
    if (!jobs || jobs.length < 1) {
      return;
    }

    var header = (
        <tr>
          <td id="buttoncol"></td>
          <td id="croncol">Cron time</td>
          <td>Workload</td>
          <td>Created</td>
          <td>Backend</td>
          <td id="buttoncol"></td>
        </tr>
    );

    var rows = jobs.map(function(job) {
      return <CronRow key={job.ID} job={job} getJobDetails={this.getJobDetails}/>;
    }.bind(this));

    return (
        <div>
          <p id="active">{title} ({rows.length})</p>
          <Table striped bordered>
            <thead>{header}</thead>
            <tbody>{rows}</tbody>
          </Table>
        </div>
      );
  },

  cronUsage: function() {
    var msg = "Note: To reduce errors, direct modifications aren't currently supported. Please instead deactivate the old job and add a new.";
    return (
      <Alert bsStyle="info">{msg}</Alert>
    );
  },

  render: function() {
    var active_jobs = (this.state.jobs || []).filter(function (val) {
        return val.IsActive;
    });
    var inactive_jobs = (this.state.jobs || []).filter(function (val) {
        return !val.IsActive;
    });

    return (
      <div className="job-details">
        <p id="job-name">{this.state.function}</p>
        {this.cronUsage()}
        <AddForm function={this.state.function} getJobDetails={this.getJobDetails}/>
        {this.displayRows("Active Jobs", active_jobs)}
        {this.displayRows("Inactive Jobs", inactive_jobs)}
      </div>
    );
  }
});
