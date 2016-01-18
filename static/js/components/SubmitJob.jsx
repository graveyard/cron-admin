var SubmitJob = React.createClass({

  getInitialState: function() {
    return {submit_success: null};
  },

  formSubmit: function(e) {
    e.preventDefault();
    var func = this.refs.function.getInputDOMNode().value;
    var workload = this.refs.workload.getInputDOMNode().value;
    $.ajax({
      url: "/submitjob",
      type: "POST",
      data: {Function: func, Workload: workload},
      success: function(data) {
        this.setState({submit_success: true});
      }.bind(this),
      error: function(xhr, status, err) {
        this.setState({submit_success: false, err_msg: err});
      }.bind(this)
    }); 
  },

  jobAlert: function() {
    if (this.state.submit_success === null) {
      return;
    }
    var msg;
    if (this.state.submit_success) {
      msg = "Successfully submitted job";
      return (<Alert bsStyle="info">{msg}</Alert>);
    } else {
      msg = "Trouble submitting job: " + this.state.err_msg;
      return (<Alert bsStyle="danger">{msg}</Alert>);
    }
  }, 

  render: function() {
    var function_placeholder = 'Function';
    var workload_placeholder = 'Workload';
    return (
        <div className="submit-job">
          {this.jobAlert()}
          <form onSubmit={this.formSubmit} method="POST">
            <Input ref="function" type="text" placeholder={function_placeholder} required />
            <Input ref="workload" type="text" placeholder={workload_placeholder}/>
            <ButtonInput bsStyle="primary" type="submit">Submit job</ButtonInput>
          </form>
        </div>
    );
  },
});
