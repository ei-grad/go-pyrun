#include "Python.h"

PyObject* GoPyRunGetContext();
PyObject* GoPyRunFileInput(PyObject *ctx, const char *command);
PyObject* GoPyRunEvalInput(PyObject *ctx, const char *command);
void GoPyRunDecref(PyObject *obj);
