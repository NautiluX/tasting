$(function() {
  start();
});

var model = {items:[]};

var start = function(){
  reloadLoop();
  $('div.name-group').popover({
    trigger: "focus",
  })
  $('input#name').change( function(){
    history.replaceState({}, null, "?name="+$(this).val());
  });
  searchParams = new URLSearchParams(window.location.search);
  $('input#name').val(searchParams.get('name'));

}

var reloadLoop = function() {
  reload();
  setTimeout(reloadLoop, 10000);
}

var reload = function(){
  $.getJSON( "/model", function(data) {
    data.items.forEach(function(item, i){
      if (model.items.length <= i || JSON.stringify(item) != JSON.stringify(model.items[i])) {
        questions = "<div class=\"container rating-block\">";
        item.rating.forEach(function(rating, r){
          questions += "<div class=\"rating\">"
          questions += "<div>" + rating.question.name + "</div>" + renderRating(item, r) + "<div><div class=\"numrating\">" + rating.avgRating + "</div></div>";
          questions += "</div>"
        })
        questions += `<button type="button" class="btn btn-primary" data-toggle="modal" data-target="#commentModal-`+item.key+`">
 Kommentieren 
</button>`;
        questions += "</div>";
        questions += `<!-- Modal -->
<div class="modal fade" id="commentModal-`+item.key+`" tabindex="-1" role="dialog" aria-labelledby="exampleModalLabel" aria-hidden="true">
  <div class="modal-dialog" role="document">
    <div class="modal-content">
      <div class="modal-header">
        <h5 class="modal-title" id="exampleModalLabel">`+item.name+` kommentieren</h5>
        <button type="button" class="close" data-dismiss="modal" aria-label="Close">
          <span aria-hidden="true">&times;</span>
        </button>
      </div>
      <div class="modal-body">
        <textarea item-key="`+item.key+`" class="form-control" aria-label="With textarea"></textarea>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-secondary" data-dismiss="modal">Abbrechen</button>
        <button type="button" class="btn btn-primary comment-button" item-key="`+item.key+`">Absenden</button>
      </div>
    </div>
  </div>
</div>`;
        html = "<h2>" + item.name + "</h2><div class=\"container item-content\"'>"+questions + buildCommentCarousel(item)+"</div>";

        if (model.items.length <= i) {
          $("<div id=\"" +item.key+ "\" class=\"container item-container\">"+html+"</div>").appendTo("div.items");
        } else {
          $("div#"+item.key).html(html);
        }
        updateRatings(item);
      }
    });
    $('.carousel').carousel();
    $('button.comment-button').unbind();
    $('button.comment-button').click(sendComment);
    $('input.ratingradio').unbind();
    $('input.ratingradio').click(sendRating);
    $('div.masthead-brand').html(data.texts.title)
    $('div#jumbo').html(data.texts.jumbo)
    model = data;
  });
}

var updateRatings = function(item) {
  item.rating.forEach(function(rating, r){
    i=1;
    for(;i<=Math.round(rating.avgRating);i++){
      $("#"+item.key+r+"-star"+i).prop("checked", true);  
    }
    for (;i<=5;i++){
      $("#"+item.key+r+"-star"+i).prop("checked", false);  
    }
  });
  
}

var renderRating = function(item, ratingNum) {
  return `<div class="container">
        <div class="starrating risingstar d-flex justify-content-center flex-row-reverse">
            <input class="ratingradio" type="radio" id="`+item.key+ratingNum+`-star5" name="`+item.key+ratingNum+`-rating" value="5" item-key="`+item.key+`" rating-num="`+ratingNum+`"/><label for="`+item.key+ratingNum+`-star5" title="5 star"></label>
            <input class="ratingradio" type="radio" id="`+item.key+ratingNum+`-star4" name="`+item.key+ratingNum+`-rating" value="4" item-key="`+item.key+`" rating-num="`+ratingNum+`"/><label for="`+item.key+ratingNum+`-star4" title="4 star"></label>
            <input class="ratingradio" type="radio" id="`+item.key+ratingNum+`-star3" name="`+item.key+ratingNum+`-rating" value="3" item-key="`+item.key+`" rating-num="`+ratingNum+`"/><label for="`+item.key+ratingNum+`-star3" title="3 star"></label>
            <input class="ratingradio" type="radio" id="`+item.key+ratingNum+`-star2" name="`+item.key+ratingNum+`-rating" value="2" item-key="`+item.key+`" rating-num="`+ratingNum+`"/><label for="`+item.key+ratingNum+`-star2" title="2 star"></label>
            <input class="ratingradio" type="radio" id="`+item.key+ratingNum+`-star1" name="`+item.key+ratingNum+`-rating" value="1" item-key="`+item.key+`" rating-num="`+ratingNum+`"/><label for="`+item.key+ratingNum+`-star1" title="1 star"></label>
        </div>
  </div>`
}

var sendComment = function() {
  key = $(this).attr('item-key');
  $('#commentModal-'+key).modal('hide');
  $('.modal-backdrop').remove();

  name = $('input#name').val();
  if (name == "") {
    $('div.name-group').popover('show');
    return;
  }
  comment = $('textarea[item-key="'+key+'"]').val()
  obj = {
    text: comment,
    item: key,
    author: name
  };
  $.ajax({
    type: "POST",
    url: "/comment",
    data: JSON.stringify(obj),
    success: function(){
      reload();
    },
    dataType: "json",
  });
}

var sendRating = function() {
  name = $('input#name').val();
  if (name == "") {
    $('div.name-group').popover('show');
    return;
  }
  key = $(this).attr('item-key');
  ratingNum = $(this).attr('rating-num');
  value = $(this).attr('value');
  obj = {
    item: key,
    ratingNum: ~~ratingNum,
    rating: ~~value,
    author: name,
  };
  $.ajax({
    type: "POST",
    url: "/rating",
    data: JSON.stringify(obj),
    success: function(){
      reload();
    },
    dataType: "json",
  });
}

var buildCommentCarousel = function(item) {
  if (item.comments.length === 0) {
    return `<div class="comment-carousel">`;
  }

  shuffledComments = item.comments
  .map((a) => ({sort: Math.random(), value: a}))
  .sort((a, b) => a.sort - b.sort)
  .map((a) => a.value)

  return `<div id="carousel-` + item.key + `" class="carousel slide comment-carousel" data-ride="carousel">
  <div class="carousel-inner comment-carousel-inner">`+
    shuffledComments.map(function(comment, i){
      return `<div class="carousel-item` + (i==0?" active":"") +`" data-interval="` + (5000+Math.floor(Math.random() * Math.floor(5000))) + `"><div class="comment container">`
        + "<h3>" + $(`<h3>`).text(`» ` + comment.text  + ` «`).html() + "</h3>" + "&nbsp;"+ $("<p/>").text(comment.author + " zu " + item.name).html() +
        `</div></div>`;
    }).join("")
  + `
  </div>
  <a class="carousel-control-prev" href="#carousel-` + item.key + `" role="button" data-slide="prev">
    <span class="carousel-control-prev-icon" aria-hidden="true"></span>
    <span class="sr-only">Previous</span>
  </a>
  <a class="carousel-control-next" href="#carousel-` + item.key + `" role="button" data-slide="next">
    <span class="carousel-control-next-icon" aria-hidden="true"></span>
    <span class="sr-only">Next</span>
  </a>
</div>`
}
