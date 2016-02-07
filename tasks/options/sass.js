module.exports = function(config) {

  return {
    src:{
      options: {
      },
      files: {
        "<%= genDir %>/css/grafana.dark.css": "<%= srcDir %>/less/grafana.dark.scss",
      }
    }
  };
};
