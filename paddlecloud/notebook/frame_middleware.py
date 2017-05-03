class NotebookMiddleware:

    def __init__(self, get_response):
        self.get_response = get_response

    def __call__(self, request):
        response = self.get_response(request)
        response["Content-Security-Policy"] = "frame-ancestors 'self' http://notebook.paddlepaddle.org"
        response["X-Frame-Options"] = "ALLOW-FROM http://notebook.paddlepaddle.org"
        return response
