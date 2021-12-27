$(document).ready(function(){ 

  console.log("Firing");

  var selectorType = ""

  console.log("editor loaded");

  $("#butt_h1").click(function() {
    console.log("You dare click me!")

  });

  $("#butt_h2").click(function() {
    console.log("You dare click me!")

  });

  $("#butt_h3").click(function() {
    console.log("You dare click me!")

  });

  $("#butt_img").click(function(e){
    console.log("press my butt");
    showModal($('#modal'));
    e.preventDefault()
  });

  $("#teaser_butt").click(function(e){
    
    selectorType = "teaser"
    showModal($('#modal'));

    e.preventDefault()
  });

  $("#search").click(function(e) {

    e.preventDefault();

    $("#searchresult").empty();

    console.log("Search for images with name of " + $("#searchbox").val());

    var searchname = $("#searchbox").val();

    if (searchname.length > 3) {
      searchname = searchname.split(' ').join('+');

      console.log("Sending search");
      $.ajax({ 
        url: '/admin/api/search-media-tags-by-name/' + searchname, 
        type: 'get', 
        dataType: 'json',
        cache: false,
        timeout: 120000,
        error: function(error) {
            $('#searchresult').html('<div class="alert alert-error" role="alert">search failed with: ' + error + '</div>');
        },
        success: function(response){ 
          console.log("success");
          if(response && response != 0){
            if (response['status'] == "success") {
              //console.log("tags is " + response.tags);
              if (response['tags'] && response['tags'].length > 0) {
                $.each(response['tags'], function(k){
                  console.log('result: ' + response['tags'][k]['document_id'] + " - " + response['tags'][k]['small_image_url']);
                  console.log('adding ' +response['tags'][k]['small_image_url'] );
                  $('#searchresult').append('<div class="img-result"><input name="img" type="radio" data-img="/'+ response['tags'][k]['small_image_url'] + '" data-title="' + response['tags'][k]['title'] + '"value="' + response['tags'][k]['document_id'] + '"/> <img src="/' + response['tags'][k]['small_image_url'] + '" style="width: 50px; height: 50px;" /><span class="img-title">' + response['tags'][k]['title']  + '</span></div>');

                });
              }else {
                $('#searchresult').html('<div class="alert alert-warn" role="warn">No Results Found</div>');
              }
            }else{
              $('#searchresult').html('<div class="alert alert-error" role="alert">search failed with: ' + response['message'] + '</div>');
            }
          }else {
            $('#searchresult').html('<div class="alert alert-warn" role="warn">No Results Found</div>');
          }
        },
        complete: function(jqXHR, status) {
          console.log("status " + status);
        }
      });
    }else{
      $("#searchresult").html("Search String Too Short");
    }
  });

  $("#enter").click(function() {

    var radioValue = $("input[name='img']:checked"). val();
    var radioTitle = $("input[name='img']:checked").attr('data-title');
    var radioURL = $("input[name='img']:checked").attr('data-img');

    console.log(radioValue + ' ' + radioURL );
    if (selectorType == "teaser"){
      $('#inputTeaserImage').val(radioValue);
      $('#imgviewer').html('<img src="' + radioURL + '" />');
    } else {
      var cursorPos = $('#mainContent').prop('selectionStart');
      console.log(cursorPos);
      var v = $('#mainContent').val();
      if (cursorPos == 0) {
        cursorPos = v.length;
      }
      console.log(v);
      var textBefore = v.substring(0,  cursorPos);
      var textAfter  = v.substring(cursorPos, v.length);
      $('#mainContent').val(textBefore + ' <div class="load-image" data-id="' + radioValue + '">' + radioTitle + '</div>' + textAfter);
    }

    selectorType = ""

    //$('#imageSelectModal').modal('hide');
  });


  //$('#imageSelectModal').on('show.bs.modal', function (e) {
    // do something...
    // console.log('visible but yet not')
  //});

  //$('#imageSelectModal').on('hide.bs.modal', function (e) {
    // do something...
  //  console.log('hidden but not forgotten')


  //  $("#searchresult").empty();

  //  $("#searchbox").val("")
  //})

  //$('#imageSelectModal').addEventListener('shown.bs.modal', function () {
  //    console.log("showing")
  //})

});
