#include "pyrun.h"

PyObject* GoPyRunGetContext() {
    PyObject *m;
    m = PyImport_AddModule("__main__");
    if (m == NULL) {
        fprintf(stderr, "PyImport_AddModule(\"__main__\") failed!");
        return NULL;
    }
    return PyModule_GetDict(m);
}

PyObject *GoPyRunFileInput(PyObject *ctx, const char *command) {
    PyObject *v;
    v = PyRun_String(command, Py_file_input, ctx, ctx);
    if (v == NULL) {
        PyErr_Print();
        return NULL;
    }
    if (Py_FlushLine())
        PyErr_Print();
    return v;
}

PyObject* GoPyRunEvalInput(PyObject *ctx, const char *command) {
    PyObject *v = PyRun_String(command, Py_eval_input, ctx, ctx);
    if (v == NULL) {
        PyErr_Print();
        return NULL;
    }
    if (Py_FlushLine())
        PyErr_Print();
    return v;
}

void GoPyRunDecref(PyObject *obj) {
    Py_DECREF(obj);
}
