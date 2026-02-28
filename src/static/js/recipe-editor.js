// Recipe Markdown Editor with custom syntax support
// Uses EasyMDE (https://github.com/Ionaru/easy-markdown-editor)

(function() {
    'use strict';

    let editors = {};
    let popups = {};

    function customPreviewRender(plainText) {
        // First, use the default markdown renderer
        let html = EasyMDE.prototype.markdown(plainText);

        // Then process our custom @ingredient{name|quantity} syntax
        html = html.replace(
            /@ingredient\{([^|]+)\|([^}]+)\}/g,
            '<span class="ingredient" data-name="$1">$2 $1</span>'
        );

        return html;
    }

    function initEditor(textareaId, options = {}) {
        const textarea = document.getElementById(textareaId);
        if (!textarea) {
            console.error('RecipeEditor: Textarea not found:', textareaId);
            return null;
        }

        const editor = new EasyMDE({
            element: textarea,
            minHeight: options.minHeight || '300px',
            placeholder: options.placeholder || 'Write your content here...',
            spellChecker: false,
            status: false,
            toolbar: [
                'bold', 'italic', 'heading', '|',
                'quote', 'unordered-list', 'ordered-list', '|',
                'link', 'table', '|',
                'preview', 'guide'
            ],
            previewRender: customPreviewRender,
            forceSync: true,
            autofocus: false,
            tabSize: 2
        });

        editors[textareaId] = editor;

        // Set up autocomplete after a short delay to ensure DOM is ready
        setTimeout(function() {
            setupAutocomplete(editor, textareaId);
        }, 100);

        // Hide sticky actions when editor is focused on mobile
        const stickyActions = document.querySelector('.sticky-actions');
        if (stickyActions) {
            const isMobile = function() {
                return window.matchMedia('(max-width: 768px)').matches;
            };
            editor.codemirror.on('focus', function() {
                if (isMobile()) {
                    stickyActions.classList.add('hidden');
                }
            });
            editor.codemirror.on('blur', function() {
                stickyActions.classList.remove('hidden');
            });
        }

        return editor;
    }

    function setupAutocomplete(editor, editorId) {
        const popup = document.createElement('div');
        popup.className = 'autocomplete-popup';
        popup.id = 'autocomplete-' + editorId;
        popup.innerHTML = '<div class="autocomplete-list"></div>';
        document.body.appendChild(popup);
        popups[editorId] = popup;

        let currentTrigger = null;
        let debounceTimer = null;

        function checkForTriggers() {
            const markdown = editor.value();

            // Check for @ingredient{ trigger (incomplete - no closing })
            const ingredientMatch = markdown.match(/@ingredient\{([^|}]*)$/);
            if (ingredientMatch) {
                currentTrigger = 'ingredient';
                showAutocomplete(popup, ingredientMatch[1], 'ingredient', editor, editorId);
                return;
            }

            // Check for [[ trigger (incomplete - no closing ]])
            const recipeMatch = markdown.match(/\[\[([^\]|]*)$/);
            if (recipeMatch) {
                currentTrigger = 'recipe';
                showAutocomplete(popup, recipeMatch[1], 'recipe', editor, editorId);
                return;
            }

            hideAutocomplete(popup);
            currentTrigger = null;
        }

        // Listen to CodeMirror's change event
        editor.codemirror.on('change', function() {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(checkForTriggers, 150);
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
            const editorWrapper = editor.codemirror.getWrapperElement();
            if (!popup.contains(e.target) && !editorWrapper.contains(e.target)) {
                hideAutocomplete(popup);
            }
        });
    }

    async function showAutocomplete(popup, query, type, editor, editorId) {
        if (!query || query.length < 1) {
            hideAutocomplete(popup);
            return;
        }

        const endpoint = type === 'ingredient' 
            ? '/api/ingredients/search?q=' + encodeURIComponent(query)
            : '/api/recipes/search?q=' + encodeURIComponent(query);

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

            // Position popup near the editor
            const editorWrapper = editor.codemirror.getWrapperElement();
            const editorRect = editorWrapper.getBoundingClientRect();
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
        const markdown = editor.value();
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
            editor.value(newMarkdown);
            editor.codemirror.focus();
            // Move cursor to end
            editor.codemirror.setCursor(editor.codemirror.lineCount(), 0);
        }
    }

    function getEditor(textareaId) {
        return editors[textareaId];
    }

    function destroyEditor(textareaId) {
        if (editors[textareaId]) {
            editors[textareaId].toTextArea();
            delete editors[textareaId];
        }
        if (popups[textareaId]) {
            popups[textareaId].remove();
            delete popups[textareaId];
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
