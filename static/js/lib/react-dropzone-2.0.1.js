// NOTE: this is pulled from https://github.com/paramaggarwal/react-dropzone

var Dropzone =
/******/ (function(modules) { // webpackBootstrap
/******/ 	// The module cache
/******/ 	var installedModules = {};

/******/ 	// The require function
/******/ 	function __webpack_require__(moduleId) {

/******/ 		// Check if module is in cache
/******/ 		if(installedModules[moduleId])
/******/ 			return installedModules[moduleId].exports;

/******/ 		// Create a new module (and put it into the cache)
/******/ 		var module = installedModules[moduleId] = {
/******/ 			exports: {},
/******/ 			id: moduleId,
/******/ 			loaded: false
/******/ 		};

/******/ 		// Execute the module function
/******/ 		modules[moduleId].call(module.exports, module, module.exports, __webpack_require__);

/******/ 		// Flag the module as loaded
/******/ 		module.loaded = true;

/******/ 		// Return the exports of the module
/******/ 		return module.exports;
/******/ 	}


/******/ 	// expose the modules object (__webpack_modules__)
/******/ 	__webpack_require__.m = modules;

/******/ 	// expose the module cache
/******/ 	__webpack_require__.c = installedModules;

/******/ 	// __webpack_public_path__
/******/ 	__webpack_require__.p = "";

/******/ 	// Load entry module and return exports
/******/ 	return __webpack_require__(0);
/******/ })
/************************************************************************/
/******/ ([
/* 0 */
/***/ function(module, exports, __webpack_require__) {

	'use strict';

	var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

	var React = __webpack_require__(1);
	var accept = __webpack_require__(2);

	var Dropzone = React.createClass({
	  displayName: 'Dropzone',

	  getDefaultProps: function getDefaultProps() {
	    return {
	      disableClick: false,
	      multiple: true
	    };
	  },

	  getInitialState: function getInitialState() {
	    return {
	      isDragActive: false
	    };
	  },

	  propTypes: {
	    onDrop: React.PropTypes.func,
	    onDropAccepted: React.PropTypes.func,
	    onDropRejected: React.PropTypes.func,
	    onDragEnter: React.PropTypes.func,
	    onDragLeave: React.PropTypes.func,

	    style: React.PropTypes.object,
	    activeStyle: React.PropTypes.object,
	    className: React.PropTypes.string,
	    activeClassName: React.PropTypes.string,
	    rejectClassName: React.PropTypes.string,

	    disableClick: React.PropTypes.bool,
	    multiple: React.PropTypes.bool,
	    accept: React.PropTypes.string
	  },

	  allFilesAccepted: function allFilesAccepted(files) {
	    var _this = this;

	    return files.every(function (file) {
	      return accept(file, _this.props.accept);
	    });
	  },

	  onDragEnter: function onDragEnter(e) {
	    e.preventDefault();

	    // This is tricky. During the drag even the dataTransfer.files is null
	    // But Chrome implements some drag store, which is accesible via dataTransfer.items
	    var dataTransferItems = e.dataTransfer && e.dataTransfer.items ? e.dataTransfer.items : [];

	    // Now we need to convert the DataTransferList to Array
	    var itemsArray = Array.prototype.slice.call(dataTransferItems);
	    var allFilesAccepted = this.allFilesAccepted(itemsArray);

	    this.setState({
	      isDragActive: allFilesAccepted,
	      isDragReject: !allFilesAccepted
	    });

	    if (this.props.onDragEnter) {
	      this.props.onDragEnter(e);
	    }
	  },

	  onDragOver: function onDragOver(e) {
	    e.preventDefault();
	  },

	  onDragLeave: function onDragLeave(e) {
	    e.preventDefault();

	    this.setState({
	      isDragActive: false,
	      isDragReject: false
	    });

	    if (this.props.onDragLeave) {
	      this.props.onDragLeave(e);
	    }
	  },

	  onDrop: function onDrop(e) {
	    e.preventDefault();

	    this.setState({
	      isDragActive: false,
	      isDragReject: false
	    });

	    var droppedFiles = e.dataTransfer ? e.dataTransfer.files : e.target.files;
	    var max = this.props.multiple ? droppedFiles.length : 1;
	    var files = [];

	    for (var i = 0; i < max; i++) {
	      var file = droppedFiles[i];
	      file.preview = URL.createObjectURL(file);
	      files.push(file);
	    }

	    if (this.props.onDrop) {
	      this.props.onDrop(files, e);
	    }

	    if (this.allFilesAccepted(files)) {
	      if (this.props.onDropAccepted) {
	        this.props.onDropAccepted(files, e);
	      }
	    } else {
	      if (this.props.onDropRejected) {
	        this.props.onDropRejected(files, e);
	      }
	    }
	  },

	  onClick: function onClick() {
	    if (!this.props.disableClick) {
	      this.open();
	    }
	  },

	  open: function open() {
	    var fileInput = React.findDOMNode(this.refs.fileInput);
	    fileInput.value = null;
	    fileInput.click();
	  },

	  render: function render() {

	    var className;
	    if (this.props.className) {
	      className = this.props.className;
	      if (this.state.isDragActive) {
	        className += ' ' + this.props.activeClassName;
	      };
	      if (this.state.isDragReject) {
	        className += ' ' + this.props.rejectClassName;
	      };
	    };

	    var style, activeStyle;
	    if (this.props.style || this.props.activeStyle) {
	      if (this.props.style) {
	        style = this.props.style;
	      }
	      if (this.props.activeStyle) {
	        activeStyle = this.props.activeStyle;
	      }
	    } else if (!className) {
	      style = {
	        width: 200,
	        height: 200,
	        borderWidth: 2,
	        borderColor: '#666',
	        borderStyle: 'dashed',
	        borderRadius: 5
	      };
	      activeStyle = {
	        borderStyle: 'solid',
	        backgroundColor: '#eee'
	      };
	    }

	    var appliedStyle;
	    if (activeStyle && this.state.isDragActive) {
	      appliedStyle = _extends({}, style, activeStyle);
	    } else {
	      appliedStyle = _extends({}, style);
	    };

	    return React.createElement(
	      'div',
	      {
	        className: className,
	        style: appliedStyle,
	        onClick: this.onClick,
	        onDragEnter: this.onDragEnter,
	        onDragOver: this.onDragOver,
	        onDragLeave: this.onDragLeave,
	        onDrop: this.onDrop
	      },
	      this.props.children,
	      React.createElement('input', {
	        type: 'file',
	        ref: 'fileInput',
	        style: { display: 'none' },
	        multiple: this.props.multiple,
	        accept: this.props.accept,
	        onChange: this.onDrop
	      })
	    );
	  }

	});

	module.exports = Dropzone;


/***/ },
/* 1 */
/***/ function(module, exports) {

	module.exports = React;

/***/ },
/* 2 */
/***/ function(module, exports) {

	/**
	 * Check if the provided file type should be accepted by the input with accept attribute.
	 * https://developer.mozilla.org/en-US/docs/Web/HTML/Element/Input#attr-accept
	 *
	 * Borrowed from https://github.com/enyo/dropzone
	 *
	 * @param file {File} https://developer.mozilla.org/en-US/docs/Web/API/File
	 * @param acceptedFiles {string}
	 * @returns {boolean}
	 */

	'use strict';

	exports.__esModule = true;

	exports['default'] = function (file, acceptedFiles) {
	    if (acceptedFiles) {
	        var _ret = (function () {
	            var acceptedFilesArray = acceptedFiles.split(',');
	            var mimeType = file.type;
	            var baseMimeType = mimeType.replace(/\/.*$/, '');

	            return {
	                v: acceptedFilesArray.some(function (type) {
	                    var validType = type.trim();
	                    if (validType.charAt(0) === '.') {
	                        return file.name.toLowerCase().indexOf(validType.toLowerCase(), file.name.length - validType.length) !== -1;
	                    } else if (/\/\*$/.test(validType)) {
	                        // This is something like a image/* mime type
	                        return baseMimeType === validType.replace(/\/.*$/, '');
	                    }
	                    return mimeType === validType;
	                })
	            };
	        })();

	        if (typeof _ret === 'object') return _ret.v;
	    }
	    return true;
	};

	module.exports = exports['default'];

/***/ }
/******/ ]);
