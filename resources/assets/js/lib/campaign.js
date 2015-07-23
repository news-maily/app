/**
 * Created by filip on 22.7.15.
 */
var Campaign = function () {

    var url = url_base + '/api/campaign';

    this.all = function (page) {
        return $.ajax({
            type: 'GET',
            url: url,
            data: {
                page: page
            },
            dataType: 'json'
        });
    };

    this.get = function (id) {
        return $.ajax({
            type: 'GET',
            url: url + '/' + id,
            dataType: 'json'
        });
    };

    this.delete = function (id) {
        return $.ajax({
            type: 'DELETE',
            url: url + '/' + id,
            dataType: 'json'
        });
    };

    this.create = function (data) {
        return $.ajax({
            type: 'POST',
            url: url,
            data: {
                name: data.name,
                subject: data.subject,
                from_name: data.from_name,
                from_email: data.from_email,
                status: data.status,
                template_id: data.template_id
            },
            dataType: 'json'
        });
    }
};

module.exports = Campaign;