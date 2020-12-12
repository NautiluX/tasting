$(function() {
  start();
});

var model = {items:[]};

var start = function(){
  reload();
}

var reload = function(){
  $.getJSON( "/model", function(data) {
    data.items.forEach(function(item, i){
      if (model.items.length <= i || JSON.stringify(item) != JSON.stringify(model.items[i])) {
        questions = "<div class=\"container rating-block\">";
        item.rating.forEach(function(rating){
          questions += "<div class=\"rating\">"
          questions += "<div>" + rating.question.name + "</div><div>" + rating.avgRating + "</div>";
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
          $("<div id=\"" +item.key+ "\">"+html+"</div>").appendTo("div.items");
        } else {
          $("div#"+item.key).html(html);
        }
      }
    });
    $('.carousel').carousel();
    $('button.comment-button').unbind();
    $('button.comment-button').click(sendComment);
    setTimeout(reload, 10000);
    model = data;
  });
}

var sendComment = function() {
  key = $(this).attr('item-key');
  comment = $('textarea[item-key="'+key+'"]').val()
  obj = {
    text: comment,
    item: key,
    author: "me"
  };
  $.post( "/comment", reload);
  $.ajax({
    type: "POST",
    url: "/comment",
    data: JSON.stringify(obj),
    success: reload,
    dataType: "json",
  });
  $('#commentModal-'+key).modal('hide')
}

var buildCommentCarousel = function(item) {

  return `<div id="carousel-` + item.key + `" class="carousel slide comment-carousel" data-ride="carousel">
  <div class="carousel-inner comment-carousel-inner">`+
    item.comments.map(function(comment, i){
      return `<div class="carousel-item` + (i==0?" active":"") +`" data-interval="` + (5000+Math.floor(Math.random() * Math.floor(5000))) + `"><div class="comment container">`
        + `<h3>&raquo; ` + comment.text  + ` &laquo;</h3> ` + comment.author + " zu " + item.name +
        `</div></div>`;
    }).join("")
  + `
  </div>
  <a class="carousel-control-prev" href="#carouselExampleIndicators" role="button" data-slide="prev">
    <span class="carousel-control-prev-icon" aria-hidden="true"></span>
    <span class="sr-only">Previous</span>
  </a>
  <a class="carousel-control-next" href="#carouselExampleIndicators" role="button" data-slide="next">
    <span class="carousel-control-next-icon" aria-hidden="true"></span>
    <span class="sr-only">Next</span>
  </a>
</div>`
}
