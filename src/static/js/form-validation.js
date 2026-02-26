const FormValidation = {
    customValidators: {},

    addValidator(fieldId, validatorFn) {
        this.customValidators[fieldId] = validatorFn;
    },

    init(formSelector) {
        const form = document.querySelector(formSelector);
        if (!form) return;

        form.setAttribute('novalidate', '');

        form.addEventListener('submit', (e) => {
            this.clearErrors(form);

            if (!this.validateForm(form)) {
                e.preventDefault();
                e.stopPropagation();
            }
        });

        form.querySelectorAll('input, textarea, select').forEach(field => {
            field.addEventListener('blur', () => {
                if (field.classList.contains('invalid')) {
                    this.validateField(field);
                }
            });

            field.addEventListener('input', () => {
                if (field.classList.contains('invalid')) {
                    this.validateField(field);
                }
            });
        });
    },

    validateForm(form) {
        let isValid = true;
        const fields = form.querySelectorAll('input, textarea, select');

        fields.forEach(field => {
            if (!this.validateField(field)) {
                isValid = false;
            }
        });

        if (!isValid) {
            const firstError = form.querySelector('.form-group.has-error');
            if (firstError) {
                firstError.scrollIntoView({ behavior: 'smooth', block: 'center' });
            }
        }

        return isValid;
    },

    validateField(field) {
        const formGroup = field.closest('.form-group');
        if (!formGroup) return true;

        this.clearFieldError(formGroup, field);

        if (!field.checkValidity()) {
            this.showFieldError(formGroup, field, this.getErrorMessage(field));
            return false;
        }

        const customValidator = this.customValidators[field.id];
        if (customValidator) {
            const error = customValidator(field);
            if (error) {
                this.showFieldError(formGroup, field, error);
                return false;
            }
        }

        return true;
    },

    getErrorMessage(field) {
        const validity = field.validity;
        const fieldName = this.getFieldName(field);

        if (validity.valueMissing) {
            return `${fieldName} is required`;
        }
        if (validity.typeMismatch) {
            if (field.type === 'email') {
                return 'Please enter a valid email address';
            }
            if (field.type === 'url') {
                return 'Please enter a valid URL';
            }
            return `Please enter a valid ${fieldName.toLowerCase()}`;
        }
        if (validity.tooShort) {
            return `${fieldName} must be at least ${field.minLength} characters`;
        }
        if (validity.tooLong) {
            return `${fieldName} must be no more than ${field.maxLength} characters`;
        }
        if (validity.rangeUnderflow) {
            return `${fieldName} must be at least ${field.min}`;
        }
        if (validity.rangeOverflow) {
            return `${fieldName} must be no more than ${field.max}`;
        }
        if (validity.patternMismatch) {
            return field.title || `Please enter a valid ${fieldName.toLowerCase()}`;
        }

        return `Please enter a valid ${fieldName.toLowerCase()}`;
    },

    getFieldName(field) {
        const label = field.closest('.form-group')?.querySelector('label');
        if (label) {
            return label.textContent.replace(/\s*\*\s*$/, '').trim();
        }
        return field.name || field.id || 'This field';
    },

    showFieldError(formGroup, field, message) {
        formGroup.classList.add('has-error');
        field.classList.add('invalid');

        let errorEl = formGroup.querySelector('.field-error');
        if (!errorEl) {
            errorEl = document.createElement('div');
            errorEl.className = 'field-error';
            const input = formGroup.querySelector('input, textarea, select, .toastui-editor-defaultUI');
            if (input) {
                input.parentNode.insertBefore(errorEl, input.nextSibling);
            } else {
                formGroup.appendChild(errorEl);
            }
        }

        errorEl.textContent = message;
        errorEl.classList.add('visible');
    },

    clearFieldError(formGroup, field) {
        formGroup.classList.remove('has-error');
        field.classList.remove('invalid');

        const errorEl = formGroup.querySelector('.field-error');
        if (errorEl) {
            errorEl.classList.remove('visible');
        }
    },

    clearErrors(form) {
        form.querySelectorAll('.form-group.has-error').forEach(group => {
            group.classList.remove('has-error');
        });
        form.querySelectorAll('.invalid').forEach(field => {
            field.classList.remove('invalid');
        });
        form.querySelectorAll('.field-error.visible').forEach(error => {
            error.classList.remove('visible');
        });
    }
};
