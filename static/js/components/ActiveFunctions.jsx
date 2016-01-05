var ActiveFunction = React.createClass({
  
  onSelected: function() {
    if (this.props.onFunctionSelected) {
      this.props.onFunctionSelected(this.props.func)
    }
  },
  
  render: function() {
      return (
        <div onClick={this.onSelected}>
        <tr>
          <td> {this.props.func} </td>
        </tr>
        </div>
      )
  }
});

var ActiveFunctionsList = React.createClass({
  render: function() {
    var FunctionItems = (this.props.functions.sort() || []).map(function(func) {
        return (
        <ActiveFunction
          key={func}
          func={func}
          onFunctionSelected={this.props.onFunctionClicked}
        />
        );
      }, this);
    return (
      <tbody>
      {FunctionItems}
      </tbody>
    )
  }
});

var ActiveFunctions = React.createClass({
  getInitialState: function() {
    this.getActiveFunctions()
    return {};
  },

  getActiveFunctions: function() {
    $.ajax({
      type: "GET",
      url: "/active-functions",
      success: function(data) {
        this.setState({functions: data})
      }.bind(this),
      error : function(xhr, status, err) {
        api_err = xhr.responsJSON.error
        console.log("Got api error :" + api_err)
      }.bind(this)
    });
  },

  onFunctionClicked: function(func) {
    this.props.navigate("jobdetails", func)
  },

  addFunction: function(e) {
    e.preventDefault()
    func = this.refs.functionName.getInputDOMNode().value
    this.props.navigate("jobdetails", func)
  },

  render: function() {
    if (!this.state.functions) {
      return (
        <div> Loading functions </div>
      )
    }

    return (
      <div>
        <div className="search-body">
          <form onSubmit={this.addFunction} method="POST">
            <Input ref="functionName" name="functionName" type="text" placeholder="Search for or add a new job by function"/>
          </form>
        </div>
        <div className="search-body">
          <Table striped bordered>
            <ActiveFunctionsList 
              functions={this.state.functions}
              onFunctionClicked={this.onFunctionClicked}
            />
          </Table>
        </div>
      </div>
    )
  }
});
