import React, { useState, useRef } from 'react';
import PropTypes from 'prop-types';
import { ControlledModal } from '../Modal/ControlledModal';
import { Button } from 'fundamental-react/Button';
import LuigiClient from '@luigi-project/client';
import { useMutationObserver } from 'react-shared';
import { SEND_NOTIFICATION } from 'gql';
import { useMutation } from 'react-apollo';

const ModalWithForm = ({
  performRefetch,
  title,
  button,
  confirmText,
  initialIsValid,
  renderForm,
  modalClassName,
}) => {
  const [isOpen, setOpen] = useState(false);
  const [isValid, setValid] = useState(initialIsValid);
  const [customValid, setCustomValid] = useState(true);
  const [sendNotification] = useMutation(SEND_NOTIFICATION);
  const formElementRef = useRef(null);

  const handleFormChanged = e => {
    setValid(formElementRef.current.checkValidity());
    if (typeof e.target.reportValidity === 'function') {
      // for IE
      e.target.reportValidity();
    }
  };

  useMutationObserver(formElementRef, () => {
    if (formElementRef.current) {
      handleFormChanged({ target: formElementRef.current });
    }
  });

  const setOpenStatus = status => {
    if (status) {
      LuigiClient.uxManager().addBackdrop();
    } else {
      LuigiClient.uxManager().removeBackdrop();
    }
    setOpen(status);
  };

  const handleFormError = (title, message) => {
    sendNotification({
      variables: {
        content: message,
        title: title,
        color: '#BB0000',
        icon: 'decline',
      },
    });
  };

  const handleFormSuccess = (title, message) => {
    sendNotification({
      variables: {
        content: message,
        title: title,
        color: '#107E3E',
        icon: 'accept',
      },
    });

    performRefetch();
  };

  const onConfirm = () => {
    const form = formElementRef.current;
    if (
      typeof form.reportValidity === 'function'
        ? form.reportValidity()
        : form.checkValidity() // IE workaround; HTML validation tooltips won't be visible
    ) {
      setOpenStatus(false);
      form.dispatchEvent(new Event('submit'));
    }
  };

  const actions = (
    <>
      <Button
        option="light"
        onClick={() => setOpenStatus(false)}
        className="fd-has-margin-right-s"
      >
        Cancel
      </Button>
      <Button
        aria-disabled={!isValid || !customValid}
        onClick={onConfirm}
        option="emphasized"
      >
        {confirmText}
      </Button>
    </>
  );

  const modalOpeningComponent = (
    <Button
      glyph={button.glyph || null}
      option={button.option}
      onClick={() => setOpenStatus(true)}
    >
      {button.text}
    </Button>
  );

  return (
    <ControlledModal
      modalOpeningComponent={modalOpeningComponent}
      modalClassName={modalClassName}
      show={isOpen}
      actions={actions}
      title={title}
      onClose={() => setOpenStatus(false)}
    >
      {renderForm({
        formElementRef,
        isValid,
        setCustomValid: isValid => {
          // revalidate rest of the form
          setValid(formElementRef.current.checkValidity());
          setCustomValid(isValid);
        },
        onChange: handleFormChanged,
        onError: handleFormError,
        onCompleted: handleFormSuccess,
      })}
    </ControlledModal>
  );
};

ModalWithForm.propTypes = {
  performRefetch: PropTypes.func.isRequired,
  sendNotification: PropTypes.func.isRequired,
  title: PropTypes.string.isRequired,
  confirmText: PropTypes.string.isRequired,
  initialIsValid: PropTypes.bool,
  button: PropTypes.exact({
    text: PropTypes.string.isRequired,
    glyph: PropTypes.string,
    option: PropTypes.oneOf(['emphasized', 'light']),
  }).isRequired,
  modalClassName: PropTypes.string,
  renderForm: PropTypes.func.isRequired,
};
ModalWithForm.defaultProps = {
  sendNotification: () => {},
  performRefetch: () => {},
  initialIsValid: false,
};

export default ModalWithForm;
