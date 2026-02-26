// Recipe Markdown Editor with custom syntax support
// Depends on TOAST UI Editor loaded via CDN

(function() {
    'use strict';

    let editors = {};
    let popups = {};

    function initEditor(elementId, options = {}) {
        const element = document.getElementById(elementId);
        if (!element) {
            console.error('RecipeEditor: Element not found:', elementId);
            return null;
        }

        const hiddenInput = document.getElementById(options.hiddenInputId);
        const initialValue = hiddenInput ? hiddenInput.value : '';

        const editor = new toastui.Editor({
            el: element,
            height: options.height || '400px',
            initialEditType: 'markdown',
            previewStyle: options.previewStyle || 'vertical',
            initialValue: initialValue,
            placeholder: options.placeholder || 'Write your content here...',
            toolbarItems: [
                ['heading', 'bold', 'italic', 'strike'],
                ['hr', 'quote'],
                ['ul', 'ol'],
                ['table', 'link'],
                ['code', 'codeblock']
            ],
            hooks: {
                addImageBlobHook: function(blob, callback) {
                    alert('Image uploads are not supported in this editor.');
                    return false;
                }
            },
            customHTMLRenderer: {
                text(node) {
                    const content = node.literal;
                    if (!content) return { type: 'text', content: '' };

                    // TODO: this could be a link and/or some hover functionality later on
                    const processed = content.replace(
                        /@ingredient\{([^|]+)\|([^}]+)\}/g,
                        '<span class="ingredient" data-name="$1">$2 $1</span>'
                    );

                    if (processed !== content) {
                        return { type: 'html', content: processed };
                    }
                    return { type: 'text', content: content };
                }
            }
        });

        // Sync to hidden input on change
        if (hiddenInput) {
            editor.on('change', function() {
                hiddenInput.value = editor.getMarkdown();
            });
        }

        editors[elementId] = editor;

        // Set up autocomplete after a short delay to ensure DOM is ready
        // TODO: after a short delay sounds like it's going to cause problems at some point in time...
        setTimeout(function() {
            setupAutocomplete(editor, element, elementId);
        }, 100);

        return editor;
    }

    function setupAutocomplete(editor, containerEl, editorId) {
        // Create popup
        const popup = document.createElement('div');
        popup.className = 'autocomplete-popup';
        popup.id = 'autocomplete-' + editorId;
        popup.innerHTML = '<div class="autocomplete-list"></div>';
        document.body.appendChild(popup);
        popups[editorId] = popup;

        let currentTrigger = null;
        // Delays API requests until user stops typing (100-150ms) to avoid request spam
        let debounceTimer = null;

        // Find the actual editable element - TOAST UI uses contenteditable divs
        const mdContainer = containerEl.querySelector('.toastui-editor-md-container');
        if (!mdContainer) {
            console.error('RecipeEditor: Could not find markdown container');
            return;
        }

        function checkForTriggers() {
            const markdown = editor.getMarkdown();

            // Check for @ingredient{ trigger (incomplete - no closing })
            const ingredientMatch = markdown.match(/@ingredient\{([^|}]*)$/);
            if (ingredientMatch) {
                currentTrigger = 'ingredient';
                showAutocomplete(popup, ingredientMatch[1], 'ingredient', containerEl, editor, editorId);
                return;
            }

            // Check for [[ trigger (incomplete - no closing ]])
            const recipeMatch = markdown.match(/\[\[([^\]|]*)$/);
            if (recipeMatch) {
                currentTrigger = 'recipe';
                showAutocomplete(popup, recipeMatch[1], 'recipe', containerEl, editor, editorId);
                return;
            }

            hideAutocomplete(popup);
            currentTrigger = null;
        }

        // Listen to the editor's change event
        editor.on('change', function() {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(checkForTriggers, 150);
        });

        // Also use input event on the container for more immediate feedback
        mdContainer.addEventListener('input', function() {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(checkForTriggers, 100);
        });

        // Handle keyboard navigation
        document.addEventListener('keydown', function(event) {
            if (!popup.classList.contains('visible')) return;

            const items = popup.querySelectorAll('.autocomplete-item');
            if (items.length === 0) return;

            const selected = popup.querySelector('.autocomplete-item.selected');
            let selectedIndex = Array.from(items).indexOf(selected);

            switch (event.key) {
                case 'ArrowDown':
                    event.preventDefault();
                    event.stopPropagation();
                    if (selected) selected.classList.remove('selected');
                    selectedIndex = (selectedIndex + 1) % items.length;
                    items[selectedIndex]?.classList.add('selected');
                    items[selectedIndex]?.scrollIntoView({ block: 'nearest' });
                    break;
                case 'ArrowUp':
                    event.preventDefault();
                    event.stopPropagation();
                    if (selected) selected.classList.remove('selected');
                    selectedIndex = selectedIndex <= 0 ? items.length - 1 : selectedIndex - 1;
                    items[selectedIndex]?.classList.add('selected');
                    items[selectedIndex]?.scrollIntoView({ block: 'nearest' });
                    break;
                case 'Enter':
                case 'Tab':
                    if (selected && popup.classList.contains('visible')) {
                        event.preventDefault();
                        event.stopPropagation();
                        insertSelection(editor, currentTrigger, selected.dataset.value);
                        hideAutocomplete(popup);
                    }
                    break;
                case 'Escape':
                    if (popup.classList.contains('visible')) {
                        event.preventDefault();
                        hideAutocomplete(popup);
                    }
                    break;
            }
        }, true);

        // Close on click outside
        document.addEventListener('click', function(e) {
            if (!popup.contains(e.target) && !containerEl.contains(e.target)) {
                hideAutocomplete(popup);
            }
        });
    }

    async function showAutocomplete(popup, query, type, containerEl, editor, editorId) {
        // Need at least 1 character to search
        if (!query || query.length < 1) {
            hideAutocomplete(popup);
            return;
        }

        const endpoint = type === 'ingredient' 
            ? '/api/ingredients/search?q=' + encodeURIComponent(query)
            : '/api/recipes/search?q=' + encodeURIComponent(query);

        console.log('RecipeEditor: Fetching', endpoint);

        try {
            const response = await fetch(endpoint);
            
            if (!response.ok) {
                console.error('RecipeEditor: Search failed:', response.status);
                hideAutocomplete(popup);
                return;
            }
            
            const results = await response.json();
            
            if (!results || results.length === 0) {
                hideAutocomplete(popup);
                return;
            }

            const list = popup.querySelector('.autocomplete-list');
            list.innerHTML = '';

            results.forEach((item, index) => {
                const div = document.createElement('div');
                div.className = 'autocomplete-item' + (index === 0 ? ' selected' : '');
                
                if (type === 'ingredient') {
                    div.textContent = item;
                    div.dataset.value = item;
                } else {
                    div.textContent = item.Title;
                    div.dataset.value = item.Title;
                }

                div.addEventListener('click', function(e) {
                    e.preventDefault();
                    e.stopPropagation();
                    insertSelection(editor, type, this.dataset.value);
                    hideAutocomplete(popup);
                });

                div.addEventListener('mouseenter', function() {
                    list.querySelectorAll('.autocomplete-item').forEach(i => i.classList.remove('selected'));
                    this.classList.add('selected');
                });

                list.appendChild(div);
            });

            // Position popup
            const editorRect = containerEl.getBoundingClientRect();
            popup.style.left = (editorRect.left + 20) + 'px';
            popup.style.top = (editorRect.top + 60) + 'px';
            popup.classList.add('visible');


        } catch (error) {
            console.error('RecipeEditor: Fetch error:', error);
            hideAutocomplete(popup);
        }
    }

    function hideAutocomplete(popup) {
        popup.classList.remove('visible');
    }

    function insertSelection(editor, type, value) {
        const markdown = editor.getMarkdown();
        let newMarkdown = markdown;

        if (type === 'ingredient') {
            newMarkdown = markdown.replace(
                /@ingredient\{([^|}]*)$/,
                '@ingredient{' + value + '|}'
            );
        } else if (type === 'recipe') {
            newMarkdown = markdown.replace(
                /\[\[([^\]|]*)$/,
                '[[' + value + ']]'
            );
        }

        if (newMarkdown !== markdown) {
            editor.setMarkdown(newMarkdown);
            editor.focus();
        }
    }

    function getEditor(elementId) {
        return editors[elementId];
    }

    function destroyEditor(elementId) {
        if (editors[elementId]) {
            editors[elementId].destroy();
            delete editors[elementId];
        }
        if (popups[elementId]) {
            popups[elementId].remove();
            delete popups[elementId];
        }
    }

    function setupFormShortcuts() {
        const form = document.getElementById('recipe-form');
        if (!form) return;

        document.addEventListener('keydown', function(e) {
            if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
                e.preventDefault();
                form.requestSubmit();
            }
        });
    }

    document.addEventListener('DOMContentLoaded', setupFormShortcuts);

    window.RecipeEditor = {
        init: initEditor,
        get: getEditor,
        destroy: destroyEditor
    };
})();
