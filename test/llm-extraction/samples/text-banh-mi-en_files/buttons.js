(function ($) {
	$(
		function () {

			$('h2:Contains("Watch how to make it")').attr('id', 'jump-watch');
			$('.single h2:Contains("Life of Dozer")').attr('id', 'jump-dozer');
			$('.single h2:Contains("In memory of Dozer")').attr('id', 'jump-dozer');
			$('.single h2:Contains("In Memory of Dozer")').attr('id', 'jump-dozer');
			$('.single h2:Contains("In Memory Of Dozer")').attr('id', 'jump-dozer');
			$('.single h2:Contains("Remembering Dozer")').attr('id', 'jump-dozer');
			$('.wprm-recipe-name').attr('id', 'jump-recipes');

			$( '.wp-block-rte-accordion' ).each( function ( index ) {
				$( this ).attr( 'id', 'faq-block-' + ( index + 1 ) );
			});
		}
	);
})(jQuery);

jQuery.expr.pseudos.Contains = jQuery.expr.createPseudo( function( arg ) {
		return function ( elem ) {
			return jQuery(elem).text().toUpperCase().indexOf(arg.toUpperCase()) >= 0;
		};
	}
);
