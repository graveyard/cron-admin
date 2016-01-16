var AddForm = React.createClass({

  getInitialState: function() {
    return {clicked: false, errored: false, json_workload_checked: false};
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
    var workload = this.refs.workload.getInputDOMNode().value;

    if (!this.state.json_workload_checked && this.raiseJSONWorkloadWarning(workload)) {
      this.setState({json_workload_checked:true});
      return;
    }

    $.ajax({
      url: "/jobs",
      type: "POST",
      data: {Function: this.props.function, CronTime: crontime, Workload: workload},
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

    var crontime_placeholder = 'Cron time: (e.g.  0 13 */4 * * 1-5)';
    var workload_placeholder = 'Workload: (e.g. --task=job or {"task":"job"})';
    var button_msg = this.state.json_workload_checked ? "Yes, submit":"Submit job";
    return (
      <div>
        {this.cronAlert()}
        {this.workloadAlert()}
        <form onSubmit={this.formSubmit} method="POST">
        <Input ref="crontime" type="text" placeholder={crontime_placeholder} required />
        <Input ref="workload" type="text" placeholder={workload_placeholder}/>
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
             TimeZone: job.TimeZone},
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

  render: function() {
    var job = this.props.job;
    var display = this.props.job.IsActive ? "Deactivate":"Activate";
    var button_display = this.state.activated_clicked ? "Are you sure?":display;
    var style = this.props.job.IsActive ? "danger":"warning";
    var delete_display = this.state.delete_clicked ? "Are you sure?":"Delete Job";
    return(
      <tr>
        <td id="button"><Button bsStyle={style} onClick={this.toggle_activated_click}>{button_display}</Button></td>
        <td>{job.CronTime}</td>
        <td id="workload">{job.Workload}</td>
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
